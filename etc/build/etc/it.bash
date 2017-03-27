#!/bin/bash -e

DRIVER_IMAGE="$1"

if [[ -z ${DRIVER_IMAGE} ]] ; then
	echo "Usage: $0 <driver image>"
	exit 1
fi

PYTHON="${PYTHON:-python3}"
DOCKER="${DOCKER:-docker}"
TOOLS="bblfsh-tools"

MANIFEST=""

TESTS_DIR="tests"
SOURCES_SUFFIX="source"
NATIVE_SUFFIX="native"
UAST_SUFFIX="uast"

green() {
	echo -e "\e[32m$1\e[0m"
}

red() {
	echo -e "\e[31m$1\e[0m"
}

yellow() {
	echo -e "\e[33m$1\e[0m"
}

require() {
	if ! which &> /dev/null $1 ; then
		red "$1 not found: $2"
		exit 1
	fi
}

# check requirements
require "${TOOLS}" "install bblfsh-sdk"
require "${PYTHON}" "install Python 3 or set its path in PYTHON environment variable"
require "${DOCKER}" "install Docker or set its path in DOCKER environment variable"

# LANGUAGE is defined by manifest
eval $("${TOOLS}" manifest)

parse_native_ast() {
	#TODO: replace with actual command
	"${DOCKER}" run -v /:/code -i "${DRIVER_IMAGE}" /opt/driver/bin/driver parse-native /code/$(readlink -f $1) | "${PYTHON}" -m json.tool
}

parse_uast() {
	#TODO: replace with actual command
	"${DOCKER}" run -v /:/code -i "${DRIVER_IMAGE}" /opt/driver/bin/driver parse-uast /code/$(readlink -f $1) | "${PYTHON}" -m json.tool
}

check_result() {
	local typ="$1"
	local target="$2"
	local tmp="$3"
	if [[ ! -e ${target} ]] ; then
                mv "${tmp}" "${target}"
                yellow "\tgenerated ${target}"
        elif [[ -f ${target} ]] ; then
                if cmp --silent "${target}" "${tmp}" ; then
                        green "\t ✔ ${typ}"
                else
                        red "\t ✖ ${typ} does not match"
                        diff -ur "${target}" "${tmp}"
                fi
        else
                red "\t✖ ${target} path not valid"
        fi
        rm -f "${tmp}"
}

mkdir -p "${TESTS_DIR}"

echo -e "Scanning sources: ${TESTS_DIR}/*.${SOURCES_SUFFIX}"
for src in $(find "${TESTS_DIR}" -type f -name '*'.${SOURCES_SUFFIX}) ; do
	NAME="$(basename "${src}")"
	NATIVE_NAME="$(echo "${NAME}" | sed -e "s/${SOURCES_SUFFIX}\$/${NATIVE_SUFFIX}/g")"
	UAST_NAME="$(echo "${NAME}" | sed -e "s/${SOURCES_SUFFIX}\$/${UAST_SUFFIX}/g")"

	green "${NAME}"

	# NATIVE
	tmp="$(mktemp)"
	parse_native_ast "${src}" > "${tmp}"
	check_result "native" "${TESTS_DIR}/${NATIVE_NAME}" "${tmp}"

	# UAST
	tmp="$(mktemp)"
	parse_uast "${src}" > "${tmp}"
	check_result "uast" "${TESTS_DIR}/${UAST_NAME}" "${tmp}"
done

