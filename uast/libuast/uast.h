#ifndef UAST_H_
#define UAST_H_

#include <stdbool.h>
#include <stdint.h>
#include <stddef.h>

typedef uintptr_t NodeHandle;

typedef uintptr_t UastHandle;

typedef enum {
    NODE_NULL,
    NODE_OBJECT,
    NODE_ARRAY,
    NODE_STRING,
    NODE_INT,
    NODE_UINT,
    NODE_FLOAT,
    NODE_BOOL,
} NodeKind;

typedef struct NodeIface {
    NodeKind (*Kind)(UastHandle, NodeHandle);

    const char * (*AsString)(UastHandle, NodeHandle);
    int64_t      (*AsInt)   (UastHandle, NodeHandle);
    uint64_t     (*AsUint)  (UastHandle, NodeHandle);
    double       (*AsFloat) (UastHandle, NodeHandle);
    bool         (*AsBool)  (UastHandle, NodeHandle);

    size_t (*Size)(UastHandle, NodeHandle);

    const char * (*KeyAt)  (UastHandle, NodeHandle, size_t);
    NodeHandle   (*ValueAt)(UastHandle, NodeHandle, size_t);


    NodeHandle (*NewObject)(UastHandle, size_t size);
    NodeHandle (*NewArray) (UastHandle, size_t size);
    NodeHandle (*NewString)(UastHandle, const char * str);
    NodeHandle (*NewInt)   (UastHandle, int64_t);
    NodeHandle (*NewUint)  (UastHandle, uint64_t);
    NodeHandle (*NewFloat) (UastHandle, double);
    NodeHandle (*NewBool)  (UastHandle, bool);

    void (*SetValue)(UastHandle, NodeHandle, size_t, NodeHandle);
    void (*SetKeyValue)(UastHandle, NodeHandle, const char *, NodeHandle);

} NodeIface;

typedef enum { PRE_ORDER, POST_ORDER, LEVEL_ORDER, POSITION_ORDER } TreeOrder;

// Uast stores the general context required for library functions.
// It must be initialized with `UastNew` passing a valid implementation of the
// `NodeIface` interface.
// Once it is not used anymore, it shall be released calling `UastFree`.
typedef struct Uast {
 NodeIface  iface;
 uintptr_t  handle;
 UastHandle ctx;
 NodeHandle root;
} Uast;

// An UastIterator is used to keep the state of the current iteration over the tree.
// It's initialized with UastIteratorNew, used with UastIteratorNext and freed
// with UastIteratorFree.
typedef struct UastIterator {
  const Uast *ctx;
  TreeOrder order;
  uintptr_t handle;
} UastIterator;

#endif // UAST_H_