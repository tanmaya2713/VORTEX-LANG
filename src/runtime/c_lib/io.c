#include <stdio.h>

void vortex_print_i32(int val) {
    printf("%d", val);
}

void vortex_print_f64(double val) {
    printf("%g", val);
}

void vortex_print_bool(int val) {
    if (val) {
        printf("true");
    } else {
        printf("false");
    }
}

void vortex_print_string(const char* str) {
    printf("%s", str);
}

void vortex_print_newline(void) {
    printf("\n");
}

// Runtime initialization stub for compiled model structures
void vortex_init(void) {
    // Intentionally left blank for base engine runtime initialization
}
