sdk: '2'
go-runtime:
  version: '1.11'
native:
  image: 'debian:latest'
  static:
  - path: 'native.sh'
    dest: 'native'
  build:
    image: 'debian:latest'
    deps:
      - 'echo dependencies'
    run:
      - 'echo build'
    artifacts:
      - path: '/native/native-binary'
        dest: 'native-binary'
  test:
    run:
      - 'echo tests'