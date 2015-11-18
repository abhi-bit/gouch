#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include "collate_json.h"
#include "min_heap.h"

typedef enum CollationMode {
    Collate_Unicode,
    Collate_Raw
} CollationMode;

minHeap* initMinHeap()
{
    minHeap *hp = (minHeap *)malloc(sizeof(minHeap));
    hp->size = 0;
    return hp;
}

void swap(node *n1, node *n2)
{
    node temp = *n1;
    *n1 = *n2;
    *n2 = temp;
}

bool compare(const sized_buf *buf1, const sized_buf *buf2, CollationMode mode)
{
    size_t length = (buf1->size < buf2->size) ? buf1->size : buf2->size;

    if (mode == Collate_Unicode) {
        int res = CollateJSON(buf1, buf2, kCollateJSON_Unicode);
        if (res != 0) return true;
        else return false;
    } else {
        if (memcmp(buf1->buf, buf2->buf, length) < 0) return true;
        else return false;
    }
}

void printArray(sized_buf *arr[], int size)
{
    int i;
    for (i = 0; i < size; i++) {
        printf("%.*s\n", (int) arr[i]->size, arr[i]->buf);
    }
}

void heapify(minHeap *hp, int i)
{
    int smallest = (LCHILD(i) < hp->size &&
                    compare(&hp->elem[LCHILD(i)].data, &hp->elem[i].data, Collate_Unicode))
                    ? LCHILD(i) : i;

    if (RCHILD(i) < hp->size &&
            compare(&hp->elem[RCHILD(i)].data, &hp->elem[smallest].data, Collate_Unicode)) {
        smallest = RCHILD(i);
    }

    if (smallest != i) {
        swap(&(hp->elem[i]), &(hp->elem[smallest]));
        heapify(hp, smallest);
    }
}

void buildMinHeap(minHeap *hp, sized_buf *arr[], int size)
{
    int i;

    for (i = 0; i < size; i++) {
        if (hp->size) {
            hp->elem = realloc(hp->elem, (hp->size + 1) * sizeof(node));
        } else {
            hp->elem = malloc(sizeof(node));
        }
        node nd;
        nd.data = *arr[i];
        hp->elem[(hp->size)++] = nd;
    }

    for (i = (hp->size - 1) / 2; i >= 0; i--) {
        heapify(hp, i);
    }
}

void insertNode(minHeap *hp, node *data) {
    if (hp->size) {
        hp->elem = realloc(hp->elem, (hp->size + 1) * sizeof(node));
    } else {
        hp->elem = malloc(sizeof(node));
    }

    node nd;
    nd.data = data->data;
    nd.i = data->i;
    nd.j = data->j;

    int i = (hp->size)++;
    while (i && compare(&nd.data, &hp->elem[PARENT(i)].data, Collate_Unicode)) {
        hp->elem[i] = hp->elem[PARENT(i)];
        i = PARENT(i);
    }
    hp->elem[i] = nd;
}

void deleteNode(minHeap *hp) {
    if (hp->size) {
        hp->elem[0] = hp->elem[--(hp->size)];
        hp->elem = realloc(hp->elem, hp->size * sizeof(node));
        heapify(hp, 0);
    } else {
        printf("\nMin Heap is empty!\n");
        free(hp->elem);
    }
}

node *getDeleteMinNode(minHeap *hp) {
    node *n = (node *)malloc(sizeof(node));
    n->data = hp->elem[0].data;
    n->i = hp->elem[0].i;
    n->j = hp->elem[0].j;
    deleteNode(hp);
    return n;
}

node *getMinNode(minHeap *hp) {
    node *n = (node *)malloc(sizeof(node));
    n->data = hp->elem[0].data;
    n->i = hp->elem[0].i;
    n->j = hp->elem[0].j;
    return n;
}

node *getMaxNode(minHeap *hp, int i) {
    if (LCHILD(i) >= hp->size) {
        return &hp->elem[i];
    }

    node *l = getMaxNode(hp, LCHILD(i));
    node *r = getMaxNode(hp, RCHILD(i));

    if (compare((const sized_buf*)&(l->data), (const sized_buf*)&(r->data),
                Collate_Unicode)) {
        return l;
    } else {
        return r;
    }
}

void replaceMin(minHeap *hp, node *n) {
    hp->elem[0].data = n->data;
    hp->elem[0].i = n->i;
    hp->elem[0].j = n->j;
    heapify(hp, 0);
}

void deleteMinHeap(minHeap *hp) {
    free(hp->elem);
}

void inorderTraversal(minHeap *hp, int i) {
    if (LCHILD(i) < hp->size) {
        inorderTraversal(hp, LCHILD(i));
    }
    printf("%.*s\n", (int) hp->elem[i].data.size, hp->elem[i].data.buf);
    if (RCHILD(i) < hp->size) {
        inorderTraversal(hp, RCHILD(i));
    }
}

void preorderTraversal(minHeap *hp, int i) {
    printf("%.*s\n", (int) hp->elem[i].data.size, hp->elem[i].data.buf);
    if (LCHILD(i) < hp->size) {
        preorderTraversal(hp, LCHILD(i));
    }
    if (RCHILD(i) < hp->size) {
        preorderTraversal(hp, RCHILD(i));
    }
}

void postorderTraversal(minHeap *hp, int i) {
    if (LCHILD(i) < hp->size) {
        postorderTraversal(hp, LCHILD(i));
    }
    if (RCHILD(i) < hp->size) {
        postorderTraversal(hp, RCHILD(i));
    }
    printf("%.*s\n", (int) hp->elem[i].data.size, hp->elem[i].data.buf);
}

void levelorderTraversal(minHeap *hp) {
    int i;
    for (i = 0; i < hp->size; i++) {
        printf("%.*s\n", (int) hp->elem[i].data.size, hp->elem[i].data.buf);
    }
}
