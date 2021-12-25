# Infracost CI scripts

This folder holds the `diff.sh` file that is included in the infracost/infracost Docker image and used in the following CI integrations:
- GitLab: [infracost-gitlab-ci](https://gitlab.com/infracost/infracost-gitlab-ci), demo is at [gitlab-ci-demo](https://gitlab.com/infracost/gitlab-ci-demo).
- CircleCI: [infracost-orb](https://github.com/infracost/infracost-orb), demos are at [circleci-github-demo](https://github.com/infracost/circleci-github-demo) and [circleci-bitbucket-demo](https://bitbucket.org/infracost/circleci-bitbucket-demo).
- Bitbucket Pipelines: [infracost-bitbucket-pipeline](https://bitbucket.org/infracost/infracost-bitbucket-pipeline), demo is at [bitbucket-pipelines-demo](https://bitbucket.org/infracost/bitbucket-pipelines-demo).
- Azure DevOps Pipelines: [infracost-azure-devops](https://github.com/infracost/infracost-azure-devops), demos are at [azure-devops-repo-demo](https://dev.azure.com/infracost/base/_git/azure-devops-repo-demo) and [azure-devops-github-demo](https://github.com/infracost/azure-devops-github-demo).

This folder also holds:
- the `atlantis_diff.sh` file that is used by the [infracost-atlantis](https://github.com/infracost/infracost-atlantis/) integration, demo is at [atlantis-demo](https://github.com/infracost/atlantis-demo).
- the `jenkins_diff.sh` file that is used by the [infracost-jenkins](https://github.com/infracost/infracost-jenkins/) integration, demo is at [jenkins-demo](https://github.com/infracost/jenkins-demo).

The idea is that when we change these bash scripts, we use the demo repos to test it works. This also means you can have clone all of the repos locally inside one folder without name conflicts.

For GitHub, see the [Infracost GitHub Actions](https://github.com/infracost/actions/), which does not use the CI scripts.
