#ifndef UAST_GO_H_
#define UAST_GO_H_

#include "uast.h"

struct NodeIface uastImpl();

extern NodeKind uastKind(UastHandle, NodeHandle);

extern char * uastAsString(UastHandle, NodeHandle);
extern int64_t      uastAsInt(UastHandle, NodeHandle);
extern uint64_t     uastAsUint(UastHandle, NodeHandle);
extern double       uastAsFloat(UastHandle, NodeHandle);
extern bool         uastAsBool(UastHandle, NodeHandle);

extern size_t uastSize(UastHandle, NodeHandle);
extern char * uastKeyAt(UastHandle, NodeHandle, size_t);
extern NodeHandle   uastValueAt(UastHandle, NodeHandle, size_t);

extern NodeHandle uastNewObject(UastHandle, size_t size);
extern NodeHandle uastNewArray(UastHandle, size_t size);

extern NodeHandle uastNewString(UastHandle, char * str);
extern NodeHandle uastNewInt(UastHandle, int64_t);
extern NodeHandle uastNewUint(UastHandle, uint64_t);
extern NodeHandle uastNewFloat(UastHandle, double);
extern NodeHandle uastNewBool(UastHandle, bool);

extern void uastSetValue(UastHandle, NodeHandle, size_t, NodeHandle);
extern void uastSetKeyValue(UastHandle, NodeHandle, char *, NodeHandle);

#endif // UAST_GO_H_