#include <stdio.h>

#ifndef MERGE_H
#define MERGE_H

#define LCHILD(x) 2 * x + 1
#define RCHILD(x) 2 * x + 2
#define PARENT(x) (x - 1) / 2

#define err(s) { printf("Allocation Failure %s\n", s); exit(EXIT_FAILURE); }
#define MALLOC(s,t) if(((s) = malloc(t)) == NULL) { err("error: malloc() ");}

typedef struct _sized_buf {
    char *buf;
    size_t size;
} sized_buf;

typedef struct node {
    sized_buf *data;
    int i; //Index of array
    int j; //Index of next element to be stored from array
} node;

typedef struct minHeap {
    int size;
    node *elem;
} minHeap;

minHeap *initMinHeap();
int getSize(minHeap *hp);
void swap(node *n1, node *n2);
void heapify(minHeap *hp, int i);
void buildMinHeap(minHeap *hp, sized_buf *arr[], int size);
void insertNode(minHeap *hp, node *data);
void deleteNode(minHeap *hp);
node *getDeleteMinNode(minHeap *hp);
node *getMinNode(minHeap *hp);
node *getMaxNode(minHeap *hp, int i);
void replaceMin(minHeap *hp, node *n);
void deleteMinHeap(minHeap *hp);
void inorderTraversal(minHeap *hp, int i);
void preorderTraversal(minHeap *hp, int i);
void postorderTraversal(minHeap *hp, int i);
void levelorderTraversal(minHeap *hp);
void printArray(sized_buf *arr[], int size);
int collate_JSON(char *b1, char *b2, int s1, int s2);
#endif
