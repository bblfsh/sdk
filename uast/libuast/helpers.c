#include "uast.h"
#include "uast_go.h"

const char * _uastAsString(UastHandle ctx, NodeHandle node) {
    char* s = uastAsString(ctx, node);
    return (const char*)(s);
}

const char * _uastKeyAt(UastHandle ctx, NodeHandle node, size_t i) {
    char* s = uastKeyAt(ctx, node, i);
    return (const char*)(s);
}

NodeHandle _uastNewString(UastHandle ctx, const char * str) {
    return uastNewString(ctx, (char *)(str));
}

void _uastSetKeyValue(UastHandle ctx, NodeHandle node, const char * k, NodeHandle v) {
    uastSetKeyValue(ctx, node, (char *)(k), v);
}

struct NodeIface uastImpl(){
    NodeIface u;
	u.Kind = uastKind;
	u.AsString = _uastAsString;
	u.AsInt = uastAsInt;
	u.AsUint = uastAsUint;
	u.AsFloat = uastAsFloat;
	u.AsBool = uastAsBool;
	u.Size = uastSize;
	u.KeyAt = _uastKeyAt;
	u.ValueAt = uastValueAt;
	u.NewObject = uastNewObject;
	u.NewArray = uastNewArray;
	u.NewString = _uastNewString;
	u.NewInt = uastNewInt;
	u.NewUint = uastNewUint;
	u.NewFloat = uastNewFloat;
	u.NewBool = uastNewBool;
	u.SetValue = uastSetValue;
	u.SetKeyValue = _uastSetKeyValue;
    return u;
}


NodeKind callKind(NodeIface* iface, UastHandle ctx, NodeHandle node) {
    return iface->Kind(ctx, node);
}

const char * callAsString(const NodeIface* iface, UastHandle ctx, NodeHandle node) {
    return iface->AsString(ctx, node);
}
int64_t callAsInt(const NodeIface* iface, UastHandle ctx, NodeHandle node) {
    return iface->AsInt(ctx, node);
}
uint64_t callAsUint(const NodeIface* iface, UastHandle ctx, NodeHandle node) {
    return iface->AsUint(ctx, node);
}
double callAsFloat(const NodeIface* iface, UastHandle ctx, NodeHandle node) {
    return iface->AsFloat(ctx, node);
}
bool callAsBool(const NodeIface* iface, UastHandle ctx, NodeHandle node) {
    return iface->AsBool(ctx, node);
}

size_t callSize(const NodeIface* iface, UastHandle ctx, NodeHandle node) {
    return iface->Size(ctx, node);
}
const char * callKeyAt(const NodeIface* iface, UastHandle ctx, NodeHandle node, size_t i) {
    return iface->KeyAt(ctx, node, i);
}
NodeHandle   callValueAt(const NodeIface* iface, UastHandle ctx, NodeHandle node, size_t i) {
    return iface->ValueAt(ctx, node, i);
}

NodeHandle callNewObject(const NodeIface* iface, UastHandle ctx, size_t size) {
    return iface->NewObject(ctx, size);
}
NodeHandle callNewArray(const NodeIface* iface, UastHandle ctx, size_t size) {
    return iface->NewArray(ctx, size);
}
NodeHandle callNewString(const NodeIface* iface, UastHandle ctx, const char * v) {
    return iface->NewString(ctx, v);
}
NodeHandle callNewInt(const NodeIface* iface, UastHandle ctx, int64_t v) {
    return iface->NewInt(ctx, v);
}
NodeHandle callNewUint(const NodeIface* iface, UastHandle ctx, uint64_t v) {
    return iface->NewUint(ctx, v);
}
NodeHandle callNewFloat(const NodeIface* iface, UastHandle ctx, double v) {
    return iface->NewFloat(ctx, v);
}
NodeHandle callNewBool(const NodeIface* iface, UastHandle ctx, bool v) {
    return iface->NewBool(ctx, v);
}

void callSetValue(const NodeIface* iface, UastHandle ctx, NodeHandle node, size_t i, NodeHandle v) {
    return iface->SetValue(ctx, node, i, v);
}
void callSetKeyValue(const NodeIface* iface, UastHandle ctx, NodeHandle node, const char * k, NodeHandle v) {
    return iface->SetKeyValue(ctx, node, k, v);
}
