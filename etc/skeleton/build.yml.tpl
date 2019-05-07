sdk: '2'
go-runtime:
  version: '1.12'
native:
  # TODO: an image used as a driver runtime
  image: 'debian:latest'
  # TODO: files copied from the source to the driver image
  static:
  - path: 'native.sh'
    dest: 'native'
  build:
    # TODO: an image to build a driver
    image: 'debian:latest'
    # TODO: build system dependencies, can not use the source
    deps:
      - 'echo dependencies'
    # TODO: build steps
    run:
      - 'echo build'
# TODO: files copied from the builder to the driver image
#    artifacts:
#      - path: '/native/native-binary'
#        dest: 'native-binary'
  test:
    # TODO: native driver tests
    run:
      - 'echo tests'