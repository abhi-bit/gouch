#include <stdio.h>
#include <stdlib.h>
#include "min_heap.h"

#define count 40

char *randstring(size_t length)
{
    static char charset[] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789,.-#'?!";
    char *randomString = NULL;
    int i;

    if (length) {
        randomString = malloc(sizeof(char) * (length +1));

        if (randomString) {
            for (int n = 0; n < length; n++) {
                int key = rand() % (int)(sizeof(charset) -1);
                randomString[n] = charset[key];
            }
            randomString[length] = '\0';
        }
    }
    return randomString;
}

int main()
{
    sized_buf *arr[count];
    int i;
    minHeap *hp = initMinHeap();

    for (i = 0; i < count; i++) {
        sized_buf *b = (sized_buf*) malloc(sizeof(sized_buf));
        b->buf = randstring(20);
        b->size = 20;
        arr[i] = b;
    }

    printf("Input array:\n");
    printArray(arr, count);
    printf("\n");
    buildMinHeap(hp, arr, count);

    printf("Deleting elements one by one:\n");
    for (i = 0; i < count; i++) {
        node *temp = getDeleteMinNode(hp);
        printf("%.*s\n", (int) temp->data.size, temp->data.buf);
        free(temp);
    }

    for (i = 0; i < count; i++) {
        free(arr[i]);
    }

    return 0;
}
