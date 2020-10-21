
pipeline {
  agent {
    node {
      label 'golang112'
    }
  }
  stages {
    stage('Go Test') {
      steps {
        sh 'export GO111MODULE="on"'
        sh 'export GOPATH=/var/gopath'
        sh 'go version'
        sh 'make test'
      }
    }    
  }
}

