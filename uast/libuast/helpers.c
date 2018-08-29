#include "uast.h"
#include "uast_go.h"

#include <stdlib.h>

NodeKind _uastKind(const Uast* ctx, NodeHandle node) {
    return uastKind((Uast*)(ctx), node);
}

const char * _uastAsString(const Uast* ctx, NodeHandle node) {
    char* s = uastAsString((Uast*)(ctx), node);
    return (const char*)(s);
}

int64_t _uastAsInt(const Uast* ctx, NodeHandle node) {
    return uastAsInt((Uast*)(ctx), node);
}

uint64_t _uastAsUint(const Uast* ctx, NodeHandle node) {
    return uastAsUint((Uast*)(ctx), node);
}

double _uastAsFloat(const Uast* ctx, NodeHandle node) {
    return uastAsFloat((Uast*)(ctx), node);
}

bool _uastAsBool(const Uast* ctx, NodeHandle node) {
    return uastAsBool((Uast*)(ctx), node);
}

size_t _uastSize(const Uast* ctx, NodeHandle node) {
    return uastSize((Uast*)(ctx), node);
}

const char * _uastKeyAt(const Uast* ctx, NodeHandle node, size_t i) {
    char* s = uastKeyAt((Uast*)(ctx), node, i);
    return (const char*)(s);
}

NodeHandle _uastValueAt(const Uast* ctx, NodeHandle node, size_t i) {
    return uastValueAt((Uast*)(ctx), node, i);
}

NodeHandle _uastNewObject(const Uast* ctx, size_t size) {
    return uastNewObject((Uast*)(ctx), size);
}

NodeHandle _uastNewArray(const Uast* ctx, size_t size) {
    return uastNewArray((Uast*)(ctx), size);
}

NodeHandle _uastNewString(const Uast* ctx, const char * str) {
    return uastNewString((Uast*)(ctx), (char *)(str));
}

NodeHandle _uastNewInt(const Uast* ctx, int64_t v) {
    return uastNewInt((Uast*)(ctx), v);
}

NodeHandle _uastNewUint(const Uast* ctx, uint64_t v) {
    return uastNewUint((Uast*)(ctx), v);
}

NodeHandle _uastNewFloat(const Uast* ctx, double v) {
    return uastNewFloat((Uast*)(ctx), v);
}

NodeHandle _uastNewBool(const Uast* ctx, bool v) {
    return uastNewBool((Uast*)(ctx), v);
}

void _uastSetValue(const Uast* ctx, NodeHandle node, size_t i, NodeHandle v) {
    uastSetValue((Uast*)(ctx), node, i, v);
}

void _uastSetKeyValue(const Uast* ctx, NodeHandle node, const char * k, NodeHandle v) {
    uastSetKeyValue((Uast*)(ctx), node, (char *)(k), v);
}

struct NodeIface* uastImpl(){
    NodeIface *u = (NodeIface*)(malloc(sizeof(NodeIface)));
	u->Kind = _uastKind;
	u->AsString = _uastAsString;
	u->AsInt = _uastAsInt;
	u->AsUint = _uastAsUint;
	u->AsFloat = _uastAsFloat;
	u->AsBool = _uastAsBool;
	u->Size = _uastSize;
	u->KeyAt = _uastKeyAt;
	u->ValueAt = _uastValueAt;
	u->NewObject = _uastNewObject;
	u->NewArray = _uastNewArray;
	u->NewString = _uastNewString;
	u->NewInt = _uastNewInt;
	u->NewUint = _uastNewUint;
	u->NewFloat = _uastNewFloat;
	u->NewBool = _uastNewBool;
	u->SetValue = _uastSetValue;
	u->SetKeyValue = _uastSetKeyValue;
    return u;
}


NodeKind callKind(NodeIface* iface, const Uast* ctx, NodeHandle node) {
    return iface->Kind(ctx, node);
}

const char * callAsString(const NodeIface* iface, const Uast* ctx, NodeHandle node) {
    return iface->AsString(ctx, node);
}
int64_t callAsInt(const NodeIface* iface, const Uast* ctx, NodeHandle node) {
    return iface->AsInt(ctx, node);
}
uint64_t callAsUint(const NodeIface* iface, const Uast* ctx, NodeHandle node) {
    return iface->AsUint(ctx, node);
}
double callAsFloat(const NodeIface* iface, const Uast* ctx, NodeHandle node) {
    return iface->AsFloat(ctx, node);
}
bool callAsBool(const NodeIface* iface, const Uast* ctx, NodeHandle node) {
    return iface->AsBool(ctx, node);
}

size_t callSize(const NodeIface* iface, const Uast* ctx, NodeHandle node) {
    return iface->Size(ctx, node);
}
const char * callKeyAt(const NodeIface* iface, const Uast* ctx, NodeHandle node, size_t i) {
    return iface->KeyAt(ctx, node, i);
}
NodeHandle   callValueAt(const NodeIface* iface, const Uast* ctx, NodeHandle node, size_t i) {
    return iface->ValueAt(ctx, node, i);
}

NodeHandle callNewObject(const NodeIface* iface, const Uast* ctx, size_t size) {
    return iface->NewObject(ctx, size);
}
NodeHandle callNewArray(const NodeIface* iface, const Uast* ctx, size_t size) {
    return iface->NewArray(ctx, size);
}
NodeHandle callNewString(const NodeIface* iface, const Uast* ctx, const char * v) {
    return iface->NewString(ctx, v);
}
NodeHandle callNewInt(const NodeIface* iface, const Uast* ctx, int64_t v) {
    return iface->NewInt(ctx, v);
}
NodeHandle callNewUint(const NodeIface* iface, const Uast* ctx, uint64_t v) {
    return iface->NewUint(ctx, v);
}
NodeHandle callNewFloat(const NodeIface* iface, const Uast* ctx, double v) {
    return iface->NewFloat(ctx, v);
}
NodeHandle callNewBool(const NodeIface* iface, const Uast* ctx, bool v) {
    return iface->NewBool(ctx, v);
}

void callSetValue(const NodeIface* iface, const Uast* ctx, NodeHandle node, size_t i, NodeHandle v) {
    return iface->SetValue(ctx, node, i, v);
}
void callSetKeyValue(const NodeIface* iface, const Uast* ctx, NodeHandle node, const char * k, NodeHandle v) {
    return iface->SetKeyValue(ctx, node, k, v);
}
