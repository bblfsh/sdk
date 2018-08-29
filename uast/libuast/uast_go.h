#ifndef UAST_GO_H_
#define UAST_GO_H_

#include "uast.h"

struct NodeIface* uastImpl();

extern NodeKind uastKind(Uast*, NodeHandle);

extern char * uastAsString(Uast*, NodeHandle);
extern int64_t      uastAsInt(Uast*, NodeHandle);
extern uint64_t     uastAsUint(Uast*, NodeHandle);
extern double       uastAsFloat(Uast*, NodeHandle);
extern bool         uastAsBool(Uast*, NodeHandle);

extern size_t uastSize(Uast*, NodeHandle);
extern char * uastKeyAt(Uast*, NodeHandle, size_t);
extern NodeHandle   uastValueAt(Uast*, NodeHandle, size_t);

extern NodeHandle uastNewObject(Uast*, size_t size);
extern NodeHandle uastNewArray(Uast*, size_t size);

extern NodeHandle uastNewString(Uast*, char * str);
extern NodeHandle uastNewInt(Uast*, int64_t);
extern NodeHandle uastNewUint(Uast*, uint64_t);
extern NodeHandle uastNewFloat(Uast*, double);
extern NodeHandle uastNewBool(Uast*, bool);

extern void uastSetValue(Uast*, NodeHandle, size_t, NodeHandle);
extern void uastSetKeyValue(Uast*, NodeHandle, char *, NodeHandle);

#endif // UAST_GO_H_