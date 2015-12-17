#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>
#include <unistd.h>
#include "collate_json.h"
#include "kway_merge.h"

int timeval_subtract(struct timeval *result, struct timeval *t2, struct timeval *t1)
{
    long int diff = (t2->tv_usec + 1000000 * t2->tv_sec) - (t1->tv_usec + 1000000 * t1->tv_sec);
    result->tv_sec = diff / 1000000;
    result->tv_usec = diff % 1000000;

    return (diff<0);
}

int main()
{

    int count = 1000000, arr_count = 2;

    minHeap *hp = initMinHeap();
    node *arr;
    MALLOC(arr, arr_count * count * sizeof(node));
    int i, j, k = 0;

    printf("Merging %d arrays each of size %d\n", arr_count, count);

    for (i = 0; i < arr_count; i++) {
        char file_name[20];
        //Text files containing one element per line
        sprintf(file_name, "%d.txt", i);
        FILE *file = fopen(file_name, "r");
        char line[22];
        for (j = 0; j < count; j++) {
            fgets(line, sizeof(line), file);
            node *n;
            MALLOC(n->data, sizeof(sized_buf));
            MALLOC(n->data->buf, 22 * sizeof(char));
            sprintf(n->data->buf, "%s", line);
            n->data->size = 20;
            arr[i * count + j] = *n;
        }
    }

    struct timeval tvBegin, tvEnd, tvDiff;
    gettimeofday(&tvBegin, NULL);
    sized_buf *output = mergeKArrays(hp, (node *)arr, arr_count, count);
    gettimeofday(&tvEnd, NULL);
    timeval_subtract(&tvDiff, &tvEnd, &tvBegin);
    printf("mergeKArrays took %ld.%06ds\n", tvDiff.tv_sec, tvDiff.tv_usec);

    for (i = 0; i < arr_count; i++) {
        for (j = 0; j < count; j++) {
            free(arr[i * count + j].data->buf);
            free(arr[i * count + j].data);
        }
    }

    /*printf("\nK-Way merge dump>\n");
    node *sn = arr;
    for (i = 0; i < arr_count * count; i++) {
        printf("%.*s\n", (int) sn->data->size, sn->data->buf);
        sn++;
    }*/

    free(arr);
    free(output);
    free(hp);
    return 0;
}
