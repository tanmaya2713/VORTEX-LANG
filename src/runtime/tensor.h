#ifndef VORTEX_TENSOR_H
#define VORTEX_TENSOR_H

typedef struct {
    int* dims;
    int ndim;
    float* data;
} VortexTensor;

VortexTensor* vortex_tensor_create(int* shape, int ndim, float* data);
void vortex_tensor_add(VortexTensor* a, VortexTensor* b, VortexTensor* out);
void vortex_tensor_relu(VortexTensor* a, VortexTensor* out);
void vortex_tensor_sigmoid(VortexTensor* a, VortexTensor* out);
void vortex_matmul(VortexTensor* a, VortexTensor* b, VortexTensor* out);
void vortex_print_tensor(VortexTensor* t);

#endif
