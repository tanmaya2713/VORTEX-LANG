#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <math.h>
#include "tensor.h"

int tensor_num_elements(const int* shape, int ndim) {
    int n = 1;
    for (int i = 0; i < ndim; i++) {
        n *= shape[i];
    }
    return n;
}

VortexTensor* vortex_tensor_create(int* shape, int ndim, float* data) {
    VortexTensor* t = (VortexTensor*)malloc(sizeof(VortexTensor));
    t->ndim = ndim;
    t->dims = (int*)malloc(ndim * sizeof(int));
    memcpy(t->dims, shape, ndim * sizeof(int));
    int n = tensor_num_elements(shape, ndim);
    t->data = (float*)malloc(n * sizeof(float));
    if (data != NULL) {
        memcpy(t->data, data, n * sizeof(float));
    } else {
        for (int i = 0; i < n; i++) {
            t->data[i] = 0.0f;
        }
    }
    return t;
}

void vortex_matmul(VortexTensor* a, VortexTensor* b, VortexTensor* out) {
    int M = a->dims[0];
    int K = a->dims[1];
    int N = b->dims[1];
    int out_size = M * N;
    for (int i = 0; i < out_size; i++) {
        out->data[i] = 0.0f;
    }
    for (int i = 0; i < M; i++) {
        for (int j = 0; j < N; j++) {
            float sum = 0.0f;
            for (int k = 0; k < K; k++) {
                sum += a->data[i * K + k] * b->data[k * N + j];
            }
            out->data[i * N + j] = sum;
        }
    }
}

void vortex_tensor_add(VortexTensor* a, VortexTensor* b, VortexTensor* out) {
    int n = tensor_num_elements(a->dims, a->ndim);
    for (int i = 0; i < n; i++) {
        out->data[i] = a->data[i] + b->data[i];
    }
}

void vortex_tensor_relu(VortexTensor* a, VortexTensor* out) {
    int n = tensor_num_elements(a->dims, a->ndim);
    for (int i = 0; i < n; i++) {
        out->data[i] = a->data[i] > 0.0f ? a->data[i] : 0.0f;
    }
}

void vortex_tensor_sigmoid(VortexTensor* a, VortexTensor* out) {
    int n = tensor_num_elements(a->dims, a->ndim);
    for (int i = 0; i < n; i++) {
        out->data[i] = 1.0f / (1.0f + expf(-a->data[i]));
    }
}

void vortex_print_tensor(VortexTensor* t) {
    printf("tensor<f32>[");
    for (int i = 0; i < t->ndim; i++) {
        if (i > 0) printf(", ");
        printf("%d", t->dims[i]);
    }
    printf("]: [");
    int n = tensor_num_elements(t->dims, t->ndim);
    for (int i = 0; i < n; i++) {
        if (i > 0) printf(", ");
        printf("%g", t->data[i]);
    }
    printf("]");
}
