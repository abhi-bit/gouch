#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "collate_json.h"
#include "kway_merge.h"

char *randstring(size_t length)
{
    //static char charset[] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789,.-#'?!";
    static char charset[] = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
    char *randomString = NULL;
    int n;

    if (length) {
        randomString = malloc(sizeof(char) * (length));
        randomString[0] = '"';
        if (randomString) {
            for (n = 1; n < length - 1; n++) {
                int key = rand() % (int)(sizeof(charset) -1);
                randomString[n] = charset[key];
            }
            randomString[length - 1] = '"';
        }
    }
    return randomString;
}

static int qsortCompare(const void *p1, const void *p2)
{
    const sized_buf *buf1 = (const sized_buf*) p1;
    const sized_buf *buf2 = (const sized_buf*) p2;

    return CollateJSON(buf1, buf2, kCollateJSON_Unicode);
}

int main()
{

    int count = 4, arr_count = 2;

    node arr[arr_count][count];
    minHeap *hp = initMinHeap();
    int i, j;

    printf("Merging %d arrays each of size %d\n\n", arr_count, count);

    for (i = 0; i < arr_count; i++) {
        for (j = 0; j < count; j++) {
            node *n = (node *) malloc(sizeof(node));
            // randstring() is bad choice, there isn't any ordering in
            // consecutive generations
            n->data.buf = randstring(20);
            n->data.size = 20;
            arr[i][j] = *n;
            printf("INSERT> %d array: ", i);
            printf("%.*s\n", (int) n->data.size, n->data.buf);
        }
    }

    qsort(arr[0], count, sizeof(node), qsortCompare);
    qsort(arr[1], count, sizeof(node), qsortCompare);

    printf("\nQ-sorted arrays\n");
    for (i = 0; i < arr_count; i++) {
        for (j = 0; j < count; j++) {
            printf("%.*s\n", (int) arr[i][j].data.size, arr[i][j].data.buf);
        }
        printf("\n");
    }
    sized_buf *output = mergeKArrays(hp, (node *)arr, arr_count, count);

    printf("\nK-Way merge dump>\n");
    for (i = 0; i < arr_count * count; i++) {
        printf("%.*s\n", (int) output[i].size, output[i].buf);
    }

    free(hp);
    return 0;
}
