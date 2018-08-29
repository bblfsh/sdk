#ifndef UAST_H_
#define UAST_H_

#include <stdbool.h>
#include <stdint.h>
#include <stddef.h>

// NodeHandle is an opaque node handle that client should use to track nodes passed to libuast.
// A handle can either be a real pointer to the node, or an ID value that client assigns to the node.
typedef uintptr_t NodeHandle;

// UastHandle is an opaque UAST context handle allocated by the client. It can be used to attach additional
// information to the UAST context. Implementation may decide to ignore the UAST context handle and interpret
// NodeHandle as pointers to node objects.
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

typedef struct Uast Uast;

// NodeIface is an interface for UAST nodes that client should implement to bind to libuast.
// Each function of the interface receives a UastHandle that can be used by the client to store
// all handle-to-node mappings for this particular UAST context. See UastHandle and NodeHandle for more details.
typedef struct NodeIface {
    NodeKind (*Kind)(const Uast*, NodeHandle);

    const char * (*AsString)(const Uast*, NodeHandle);
    int64_t      (*AsInt)   (const Uast*, NodeHandle);
    uint64_t     (*AsUint)  (const Uast*, NodeHandle);
    double       (*AsFloat) (const Uast*, NodeHandle);
    bool         (*AsBool)  (const Uast*, NodeHandle);

    size_t (*Size)(const Uast*, NodeHandle);

    const char * (*KeyAt)  (const Uast*, NodeHandle, size_t);
    NodeHandle   (*ValueAt)(const Uast*, NodeHandle, size_t);


    NodeHandle (*NewObject)(const Uast*, size_t size);
    NodeHandle (*NewArray) (const Uast*, size_t size);
    NodeHandle (*NewString)(const Uast*, const char * str);
    NodeHandle (*NewInt)   (const Uast*, int64_t);
    NodeHandle (*NewUint)  (const Uast*, uint64_t);
    NodeHandle (*NewFloat) (const Uast*, double);
    NodeHandle (*NewBool)  (const Uast*, bool);

    void (*SetValue)(const Uast*, NodeHandle, size_t, NodeHandle);
    void (*SetKeyValue)(const Uast*, NodeHandle, const char *, NodeHandle);

} NodeIface;

typedef enum { PRE_ORDER, POST_ORDER, LEVEL_ORDER, POSITION_ORDER } TreeOrder;

#define UAST_CALL(u, name, ...) u->iface->name(u, __VA_ARGS__)

// Uast stores the general context required for library functions.
// It must be initialized with UastNew passing a valid implementation of the NodeIface interface.
// Once it is not used anymore, it shall be released calling `UastFree`.
typedef struct Uast {
 // iface is an implementation of the node interface that will be used for this UAST context.
 NodeIface  *iface;
 // handle is an internal UAST context handle defined by the libuast. It shouldn't be changed our used by the client.
 uintptr_t  handle;
 // ctx is an optional UAST handle that libuast will pass to each node interface function.
 // It can be used to track different UAST contexts in the client code.
 UastHandle ctx;
 // root is an optional root node handle that will be used by default if Filter, Encode and other operations.
 NodeHandle root;
} Uast;

// An UastIterator is used to keep the state of the current iteration over the tree.
// It's initialized with UastIteratorNew, used with UastIteratorNext and freed
// with UastIteratorFree.
typedef struct UastIterator {
  const Uast *ctx;
  uintptr_t handle;
} UastIterator;

typedef enum { UAST_BINARY = 0, UAST_YAML = 1 } UastFormat;

// UastLoad copies the node from a source context into the destination.
//
// Since contexts might be backed by different node interface implementations,
// this functions allows to load UAST to and from the libuast-owned memory.
static NodeHandle UastLoad(const Uast *src, NodeHandle n, const Uast *dst) {
    NodeKind kind = UAST_CALL(src, Kind, n);

    if (kind == NODE_NULL) {
        return 0;
    } else if (kind == NODE_OBJECT) {
        size_t sz = UAST_CALL(src, Size, n);

        NodeHandle m = UAST_CALL(dst, NewObject, sz);
        for (size_t i = 0; i < sz; i++) {
            const char * k = UAST_CALL(src, KeyAt, n, i);
            if (!k) {
                return 0;
            }
            NodeHandle v = UAST_CALL(src, ValueAt, n, i);
            v = UastLoad(src, v, dst);
            UAST_CALL(dst, SetKeyValue, m, k, v);
        }
        return m;
      } else if (kind == NODE_ARRAY) {
        size_t sz = UAST_CALL(src, Size, n);

        NodeHandle m = UAST_CALL(dst, NewArray, sz);
        for (size_t i = 0; i < sz; i++) {
            NodeHandle v = UAST_CALL(src, ValueAt, n, i);
            v = UastLoad(src, v, dst);
            UAST_CALL(dst, SetValue, m, i, v);
        }
        return m;
      } else if (kind == NODE_STRING) {
        return UAST_CALL(dst, NewString, UAST_CALL(src, AsString, n));
      } else if (kind == NODE_INT) {
        return UAST_CALL(dst, NewInt, UAST_CALL(src, AsInt, n));
      } else if (kind == NODE_UINT) {
        return UAST_CALL(dst, NewUint, UAST_CALL(src, AsUint, n));
      } else if (kind == NODE_FLOAT) {
        return UAST_CALL(dst, NewFloat, UAST_CALL(src, AsFloat, n));
      } else if (kind == NODE_BOOL) {
        return UAST_CALL(dst, NewBool, UAST_CALL(src, AsBool, n));
      }
      return 0;
}

#endif // UAST_H_