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
SOURCES_DIR="${TESTS_DIR}/sources"
NATIVE_DIR="${TESTS_DIR}/native"
UAST_DIR="${TESTS_DIR}/uast"

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

mkdir -p "${SOURCES_DIR}"
mkdir -p "${NATIVE_DIR}"
mkdir -p "${UAST_DIR}"

echo -e "Scanning sources in ${SOURCES_DIR}"
find "${SOURCES_DIR}" -type f | while read src ; do
	NAME="$(basename "${src}")"
	NATIVE="${NATIVE_DIR}/${NAME}.json"
	UAST="${UAST_DIR}/${NAME}.json"

	green "${NAME}"

	# NATIVE
	tmp="$(mktemp)"
	parse_native_ast "${src}" > "${tmp}"
	check_result "native" "${NATIVE}" "${tmp}"

	# UAST
	tmp="$(mktemp)"
	parse_uast "${src}" > "${tmp}"
	check_result "uast" "${UAST}" "${tmp}"
done

