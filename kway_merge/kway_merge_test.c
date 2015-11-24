#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "collate_json.h"
#include "min_heap.h"

#define count 4
#define arr_count 4

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

static int compare(const void *p1, const void *p2)
{
    const sized_buf *buf1 = (const sized_buf*) p1;
    const sized_buf *buf2 = (const sized_buf*) p2;

    return CollateJSON(buf1, buf2, kCollateJSON_Unicode);
}

sized_buf *mergeKArrays(minHeap *hp, node buf_arr[arr_count][count])
{
    sized_buf *output = (sized_buf *)malloc(sizeof(sized_buf) * arr_count * count);
    int i, j;

    for (i = 0; i < arr_count; i++) {
        node *n = (node *)malloc(sizeof(node));
        n->data = buf_arr[i][0].data;
        n->i = i;
        // j represents next element to be picked from the ith array
        n->j = 1;
        insertNode(hp, n);
    }

    for (j = 0; j < arr_count * count; j++) {
        node *root = getMinNode(hp);
        output[j] = root->data;

        if (root->j < count) {
            root->data = buf_arr[root->i][root->j].data;
            root->j += 1;
        } else {
            sized_buf data;
            data.buf = "\"z\"";
            data.size = 2;
            root->data = data;
        }

        replaceMin(hp, root);
    }
    return output;
}

int main()
{
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

    qsort(arr[0], count, sizeof(node), compare);
    qsort(arr[1], count, sizeof(node), compare);

    printf("\nQ-sorted arrays\n");
    for (i = 0; i < arr_count; i++) {
        for (j = 0; j < count; j++) {
            printf("%.*s\n", (int) arr[i][j].data.size, arr[i][j].data.buf);
        }
        printf("\n");
    }
    sized_buf *output = mergeKArrays(hp, arr);

    printf("\nK-Way merge dump>\n");
    for (i = 0; i < arr_count * count; i++) {
        printf("%.*s\n", (int) output[i].size, output[i].buf);
    }

    free(hp);
    return 0;
}
