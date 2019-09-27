pipeline {
  agent {
    kubernetes {
      label 'sdk-drivers-updater'
      defaultContainer 'sdk-drivers-updater'
      yaml """
spec:
  nodeSelector:
    srcd.host/type: jenkins-worker
  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
          - key: jenkins
            operator: In
            values:
            - slave
        topologyKey: kubernetes.io/hostname
  containers:
  - name: sdk-drivers-updater
    image: bblfsh/performance:latest
    imagePullPolicy: Always
    securityContext:
      privileged: true
    command:
    - dockerd
    tty: true
"""
    }
  }
  triggers {
    GenericTrigger(
      genericVariables: [
        [key: 'target', value: '$.target'],
        [key: 'title', value: '$.title'],
        [key: 'text', value: '$.text'],
        [key: 'sdk_version', value: '$.sdk_version'],
        [key: 'branch', value: '$.branch'],
        [key: 'commit_msg', value: '$.commit_msg'],
        [key: 'script', value: '$.script']
      ],
      token: 'update',
      causeString: 'Triggered on $target',

      printContributedVariables: true,
      printPostContent: true,

      regexpFilterText: '$target',
      regexpFilterExpression: 'master'
    )
  }
  stages {
    stage('Run updater') {
      when { branch 'master' }
      steps {
        withCredentials([usernamePassword(credentialsId: '87b3cad8-8b12-4e91-8f47-33f3d7d45620', passwordVariable: 'token', usernameVariable: 'user')]) {
          sh 'echo ${script} > /etc/script.sh ; chmod +x /etc/script.sh'
          sh 'GITHUB_TOKEN=${token} go run cmd/bblfsh-drivers-updater/update.go --script="/etc/script.sh" --sdk-version="${sdk_version}" --branch="${branch}" --title="${title}"  --text="${text}" --commit-msg="${commit_msg}" --dockerfile=true'
        }
      }
    }
  }
}
