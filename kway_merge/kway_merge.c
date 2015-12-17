#include <stdlib.h>
#include "kway_merge.h"

sized_buf *mergeKArrays(minHeap *hp, node *buf_arr, int arr_count, int count)
{
    sized_buf *output;
    MALLOC(output, sizeof(sized_buf) * arr_count * count);
    int i, j, k;
    node *z;

    for (i = 0; i < arr_count; i++) {
        node *n = (node *)malloc(sizeof(node));
        z = buf_arr + i * count;
        n->data = z->data;
        n->i = i;
        // j represents next element to be picked from the ith array
        n->j = 1;
        insertNode(hp, n);
    }

    for (k = 0; k < arr_count * count; k++) {
        node *root = getMinNode(hp);
        output[k] = *(root->data);

        if (root->j < count) {
            i = root->i;
            j = root->j;
            z = buf_arr + i * count;
            root->data = (z + j)->data;
            root->j += 1;
        } else {
            sized_buf data;
            data.buf = "\"z\"";
            data.size = 2;
            root->data = &data;
        }

        replaceMin(hp, root);
    }
    return output;
}

