# Infracost CI scripts

For GitHub, see the [Infracost GitHub Actions](https://github.com/infracost/actions/), which does not use these CI scripts.

For GitLab, see the [Infracost GitLab CI](https://gitlab.com/infracost/infracost-gitlab-ci) examples, which uses comment.sh.

## Older scripts that will be deprecated at some point

### diff.sh
This folder holds the `diff.sh` file that is included in the infracost/infracost Docker image and used in the following CI integrations:
- CircleCI: [infracost-orb](https://github.com/infracost/infracost-orb), demos are at [circleci-github-demo](https://github.com/infracost/circleci-github-demo) and [circleci-bitbucket-demo](https://bitbucket.org/infracost/circleci-bitbucket-demo).
- Bitbucket Pipelines: [infracost-bitbucket-pipeline](https://bitbucket.org/infracost/infracost-bitbucket-pipeline), demo is at [bitbucket-pipelines-demo](https://bitbucket.org/infracost/bitbucket-pipelines-demo).
- Azure DevOps Pipelines: [infracost-azure-devops](https://github.com/infracost/infracost-azure-devops), demos are at [azure-devops-repo-demo](https://dev.azure.com/infracost/base/_git/azure-devops-repo-demo) and [azure-devops-github-demo](https://github.com/infracost/azure-devops-github-demo).

### atlantis_diff.sh
The `atlantis_diff.sh` file that is used by the [infracost-atlantis](https://github.com/infracost/infracost-atlantis/) integration, demo is at [atlantis-demo](https://github.com/infracost/atlantis-demo).

### jenkins_diff.sh
The `jenkins_diff.sh` file that is used by the [infracost-jenkins](https://github.com/infracost/infracost-jenkins/) integration, demo is at [jenkins-demo](https://github.com/infracost/jenkins-demo).
