
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
        sh 'go test ./...'
        sh 'go version'
        sh 'go get -t -v ./...'
        sh 'go test -v -race -covermode=atomic ./...'
      }
    }    
  }
}

