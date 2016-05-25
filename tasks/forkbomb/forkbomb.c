#include <stdio.h>
#include <unistd.h>

// MAXPROC is equals to 100 but we have the run.sh and the forkbomb.c
// that are already running.
#define MAXPROC 98

int main(void) {
    printf("Start\n");

    int bomb_count = 0;
    while (1) {
        switch (fork()) {
            case -1:
                // An error occured
                if (bomb_count == MAXPROC) {
                    printf("Done\n");
                    return 0;
                } else {
                    return 1;
                }
                break;
            case 0:
                // We are in the child process
                // Just keep the child busy for a while
                sleep(4);
                break;
            default:
                // We are in the parent process
                bomb_count++;
                break;
        }
    }

    return 1;
}
