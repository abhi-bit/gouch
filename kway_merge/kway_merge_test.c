#include <stdio.h>
#include <stdlib.h>
#include "min_heap.h"

#define count 4
#define arr_count 2

char *randstring(size_t length)
{

    static char charset[] = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789,.-#'?!";
    char *randomString = NULL;

    if (length) {
        randomString = malloc(sizeof(char) * (length +1));

        if (randomString) {
            for (int n = 0;n < length;n++) {
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

    for (int i = 0; i < arr_count; i++) {
        node *n = (node *)malloc(sizeof(node));
        n->data = buf_arr[i][0].data;
        n->i = i;
        n->j = 1;
        insertNode(hp, n);
    }

    for (int j = 0; j < arr_count * count; j++) {
        node *root = getMinNode(hp);
        output[j] = root->data;

        if (root->j < count) {
            root->data = buf_arr[root->i][root->j].data;
            root->j += 1;
        } else {
            sized_buf data;
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

    for (int i = 0; i < arr_count; i++) {
        for (int j = 0; j < count; j++) {
            node *n = (node *) malloc(sizeof(node));
            n->data.buf = randstring(20);
            n->data.size = 20;
            arr[i][j] = *n;
            printf("Inserted: ");
            printf("%.*s\n", (int) n->data.size, n->data.buf);
        }
    }

    sized_buf *output = mergeKArrays(hp, arr);

    printf("\nArray Dump:\n");
    for (int i = 0; i < arr_count * count; i++) {
        printf("%.*s\n", (int) output[i].size, output[i].buf);
    }

    printf("Deleting nodes: \n");
    node *temp = getDeleteMinNode(hp);
    printf("%.*s\n", (int) temp->data.size, temp->data.buf);
    temp = getDeleteMinNode(hp);
    printf("%.*s\n", (int) temp->data.size, temp->data.buf);
    temp = getDeleteMinNode(hp);
    printf("%.*s\n", (int) temp->data.size, temp->data.buf);

    free(hp);
    return 0;
}
