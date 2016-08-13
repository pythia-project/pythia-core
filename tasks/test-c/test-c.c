#include<stdio.h>

#define MAX_INPUT 100

int main() {
    char line[MAX_INPUT];
    while (fgets(line, MAX_INPUT, stdin)) {
        printf("%s", line);
    }
    return 0;
}
