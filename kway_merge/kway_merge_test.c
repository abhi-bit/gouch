#include <stdio.h>
#include <stdlib.h>
#include "min_heap.h"

#define count 4
#define arr_count 4

char *randstring(size_t length)
{
    static char charset[] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789,.-#'?!";
    char *randomString = NULL;
    int n;

    if (length) {
        randomString = malloc(sizeof(char) * (length +1));

        if (randomString) {
            for (n = 0;n < length;n++) {
                int key = rand() % (int)(sizeof(charset) -1);
                randomString[n] = charset[key];
            }
            randomString[length] = '\0';
        }
    }
    return randomString;
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
            // Using "~" as marker, it being the last legible
            // element in ascii table
            data.buf = "~";
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

    sized_buf *output = mergeKArrays(hp, arr);

    printf("\nK-Way merge dump>\n");
    for (i = 0; i < arr_count * count; i++) {
        printf("%.*s\n", (int) output[i].size, output[i].buf);
    }

    free(hp);
    return 0;
}
