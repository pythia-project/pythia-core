// Copyright 2013 The Pythia Authors.
// This file is part of Pythia.
//
// Pythia is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, version 3 of the License.
//
// Pythia is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Pythia.  If not, see <http://www.gnu.org/licenses/>.

#define _GNU_SOURCE

#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <signal.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/time.h>
#include <sys/resource.h>
#include <sys/wait.h>
#include <sys/mount.h>
#include <sys/ipc.h>
#include <sys/shm.h>
#include <sys/sem.h>
#include <sys/msg.h>
#include <linux/reboot.h>

#define LOGNAME "pythia"

//! Maximum length of the disksize vm parameter.
#define DISKSIZE_MAXLEN 10

//! Per-user process limit
#define MAXPROC 100

//! User ID of the privileged non-root user
#define UID_MASTER 1

//! User ID of the unprivileged user
#define UID_WORKER 2

//! Maximum line size in /task/control.
#define CONTROL_MAXLEN 4096

//! Maximum number of arguments in a command of /task/control.
#define CONTROL_MAXARGS 100

/**
 * Shut down the virtual machine.
 */
static inline void shutdown() {
    reboot(LINUX_REBOOT_CMD_HALT);
}

/**
 * Print a log message.
 */
static inline void msg(const char *s) {
    printf("%s: %s\n", LOGNAME, s);
}

/**
 * Print a log message and shut down.
 */
static inline void msgdie(const char *s) {
    msg(s);
    shutdown();
}

/**
 * Print error using perror(3) and shut down.
 */
static inline void die() {
    perror(LOGNAME);
    shutdown();
}

/**
 * If result is non-zero, print error using perror(3) and shut down.
 */
static inline void check(int result) {
    if(result != 0)
        die();
}

/**
 * Print error using perror(3) and exit with non-zero status.
 */
static inline void childdie() {
    perror(LOGNAME);
    exit(1);
}

/**
 * If result is non-zero, print error using perror(3) and exit with non-zero
 * status.
 */
static inline void childcheck(int result) {
    if(result != 0)
        childdie();
}

/**
 * Perform cleanup after executing a step.
 * The following operations are performed.
 * - Send KILL signal to all processes, except init (this program).
 * - Release all shared memory.
 * - Release all semaphores.
 * - Release all message queues.
 */
static void cleanup() {
    int n, i, id;
    struct shminfo shminfo;
    struct shmid_ds shm;
    struct seminfo seminfo;
    struct semid_ds sem;
    union semun {
        int              val;
        struct semid_ds *buf;
        unsigned short  *array;
        struct seminfo  *__buf;
    } semun;
    struct msginfo msginfo;
    struct msqid_ds msq;

    semun.val = 0;

    // Kill processes
    kill(-1, SIGKILL);

    // Release shared memory
    n = shmctl(0, IPC_INFO, (struct shmid_ds *) &shminfo);
    if(n < 0)
        die();
    for(i = 0; i <= n; i++) {
        id = shmctl(i, SHM_STAT, &shm);
        if(id >= 0)
            shmctl(id, IPC_RMID, NULL);
    }

    // Release semaphores
    semun.__buf = &seminfo;
    n = semctl(0, 0, IPC_INFO, sem);
    if(n < 0)
        die();
    semun.buf = &sem;
    for(i = 0; i <= n; i++) {
        id = semctl(i, 0, SEM_STAT, sem);
        if(id >= 0)
            semctl(id, 0, IPC_RMID, sem);
    }

    // Release message queues
    n = msgctl(0, IPC_INFO, (struct msqid_ds *) &msginfo);
    if(n < 0)
        die();
    for(i = 0; i <= n; i++) {
        id = msgctl(i, MSG_STAT, &msq);
        if(id >= 0)
            msgctl(id, IPC_RMID, NULL);
    }
}

/**
 * Split a command line into arguments.
 *
 * We try to respect shell conventions.
 * - Arguments are separated by whitespace(s) [ \t\b\v\r\n].
 * - Whitespace can be enclosed by single (') or double quotes (").
 * - A double quote inside double quotes can be escaped by a backslash (\)
 *
 * @param[in] cmd the command line (will be modified)
 * @param[out] argv will contain the arguments (an array of at least
 *                  CONTROL_MAXARGS+1 elements)
 */
static void splitargs(char *cmd, char **argv) {
    int idx, offset;
    char quote, replacement;

    idx = 0;
    argv[idx] = NULL;
    offset = 0;
    quote = 0;

    /* Invariants:
     * - argv[idx] == NULL || argv[idx+1] == NULL
     * - argv[idx] == NULL => offset == 0 && quote == 0
     */

    while(cmd[0] != '\0') {
        switch(cmd[0]) {
        case ' ':
        case '\t':
        case '\r':
        case '\n':
            if(argv[idx] == NULL) {
                // Already reading whitespace, do nothing
            } else if(quote == 0) {
                // No quotes, end of argument
                cmd[offset] = '\0';
                idx++;
                offset = 0;
            } else {
                // Inside quotes, read as normal character
                cmd[offset] = cmd[0];
            }
            break;
        case '"':
        case '\'':
            if(argv[idx] == NULL) {
                // Was reading whitespace, start of quoted argument
                argv[idx] = cmd + 1;
                argv[idx+1] = NULL;
                quote = cmd[0];
            } else if(quote == cmd[0]) {
                // In matching quotes, end of quotes
                quote = 0;
                offset--;
            } else if(quote == 0) {
                // Outside quotes, start of quotes
                quote = cmd[0];
                offset--;
            } else {
                // Read as normal character
                cmd[offset] = cmd[0];
            }
            break;
        case '\\':
            if(quote != '\'') {
                replacement = 0;
                switch(cmd[1]) {
                case 'a': replacement = '\a'; break;
                case 'b': replacement = '\b'; break;
                case 'f': replacement = '\f'; break;
                case 'n': replacement = '\n'; break;
                case 'r': replacement = '\r'; break;
                case 't': replacement = '\t'; break;
                case 'v': replacement = '\v'; break;
                case '\\':
                case '\'':
                case '"':
                    replacement = cmd[1];
                    break;
                }
                if(replacement != 0) {
                    cmd[offset] = replacement;
                    offset--;
                    cmd++;
                    break;
                }
            }
            // else fallthrough
        default:
            if(argv[idx] == NULL) {
                // Was reading whitespace, start of argument
                argv[idx] = cmd;
                argv[idx+1] = NULL;
            }
            cmd[offset] = cmd[0];
        }
        cmd++;
        if(idx >= CONTROL_MAXARGS)
            msgdie("arguments limit exceeded");
    }
    cmd[offset] = '\0';
    if(quote != 0)
        msgdie("unbalanced quotes");
}

/**
 * Environment used when launching programs.
 */
static char *const ENVIRONMENT[] = {
    "PATH=/usr/bin:/bin",
    "LANG=C",
    "HOME=/tmp",
    NULL
};

/**
 * Handle to the /task/control file used in run_control.
 * It is defined here so launch() can close it in the child process.
 */
static FILE *fcontrol;

/**
 * Launches a program and wait for it to finish.
 *
 * If uid is not UID_MASTER, the standard input and output will be redirected
 * to /dev/null.
 *
 * If uid is UID_MASTER and the program exits with non-zero status (or an error
 * occurs during the setup of the child process), the vm will be shut down.
 *
 * The umask also depends on uid. For UID_MASTER, files will be private by
 * default. For other users, files will be public by default.
 *
 * @param cmd the command to execute (will be modified)
 * @param uid the user id that will execute the program
 */
static void launch(char *cmd, uid_t uid) {
    char *argv[CONTROL_MAXARGS+1];
    pid_t pid;
    int status;

    splitargs(cmd, argv);
    pid = fork();
    if(pid < 0)
        die();
    if(pid > 0) {
        // Parent
        wait(&status);
        if(uid == UID_MASTER &&
                (!WIFEXITED(status) || WEXITSTATUS(status) != 0))
            shutdown();
    } else {
        // Child
        childcheck(setuid(uid));
        childcheck(fclose(fcontrol));
        if(uid == UID_MASTER) {
            // Make new files private to master by default
            umask(077);
        } else {
            // Make new files public by default
            umask(000);
            // Deny access to input and output
            if(freopen("/dev/null", "r", stdin) == NULL ||
               freopen("/dev/null", "w", stdout) == NULL ||
               freopen("/dev/null", "w", stderr) == NULL)
                childdie();
        }
        execve(argv[0], argv, ENVIRONMENT);
        childdie();  // if we arrive here, there was an error launching cmd.
    }
}

/**
 * Read /task/control and execute the commands.
 * The file contains one command per line. If a line starts with '!', it will
 * be run unprivileged.
 */
static void run_control() {
    char line[CONTROL_MAXLEN+1];

    fcontrol = fopen("/task/control", "r");
    if(fcontrol == NULL)
        die();
    while(fgets(line, CONTROL_MAXLEN+1, fcontrol) != NULL) {
        if(line[0] == '!')
            launch(line + 1, UID_WORKER);
        else
            launch(line, UID_MASTER);
        cleanup();
    }
}

#define TMPFS_PARAMS "mode=777,size="

/**
 * Init entry point.
 */
int main() {
    const char *disksize;
    size_t disksize_len;
    char tmpfsdata[sizeof(TMPFS_PARAMS)+DISKSIZE_MAXLEN];
    struct rlimit rlim;

    // Print start marker
    msg("init");

    // Parse environment variables
    disksize = getenv("disksize");
    if(disksize == NULL)
        disksize = "50%";
    disksize_len = strlen(disksize);
    if(disksize_len > DISKSIZE_MAXLEN)
        msgdie("disksize parameter is too long");
    memcpy(tmpfsdata, TMPFS_PARAMS, sizeof(TMPFS_PARAMS)-1);
    memcpy(tmpfsdata+sizeof(TMPFS_PARAMS)-1, disksize, disksize_len+1);

    // Mount essential filesystems
    check(mount("proc",      "/proc", "proc",     MS_NODEV | MS_NOSUID | MS_NOEXEC, NULL));
    check(mount("sys",       "/sys",  "sysfs",    MS_NODEV | MS_NOSUID | MS_NOEXEC, NULL));
    check(mount("none",      "/tmp",  "tmpfs",    MS_NODEV | MS_NOSUID,             tmpfsdata));

    // Mount task filesystem
    check(mount("/dev/ubdb", "/task", "squashfs", MS_NODEV | MS_NOSUID | MS_RDONLY, NULL));

    // Limit the number of processes a user may create
    rlim.rlim_max = rlim.rlim_cur = MAXPROC;
    check(setrlimit(RLIMIT_NPROC, &rlim));

    // Do real work
    run_control();

    // Finish
    shutdown();
    return 0;
}

// vim:set ts=4 sw=4 et:
