---
name: Release a new version of Hubble CLI from main branch
about: A checklist for Hubble CLI release process
title: 'vX.Y.Z release'
---

- [ ] Install [`gh`](https://cli.github.com/) and authenticate with GitHub by running `gh auth login`.

- [ ] Define `NEW_RELEASE` and `PREVIOUS_RELEASE` environment variables. For
      example, if you are releasing a new Hubble CLI version based on
      `Cilium 1.16.1` with the previous release being `Cilium 1.16.0`:

      export NEW_RELEASE=1.16.1
      export PREVIOUS_RELEASE=1.16.0

- [ ] Create a release prep branch:

      git checkout main
      git pull origin main
      git switch -c pr/$USER/v$NEW_RELEASE-prep

- [ ] Check if `replace` directive in `go.mod` is in sync with `cilium/cilium`. Run:

       curl https://raw.githubusercontent.com/cilium/cilium/$NEW_RELEASE/go.mod

     and copy the `replace` directive to `go.mod` if it's out of sync.

- [ ] Update Cilium dependency:

      go get github.com/cilium/cilium@${NEW_RELEASE}
      go mod tidy && go mod vendor && go mod verify
      git add go.mod go.sum vendor

- [ ] Prepare release notes. You need to generate release notes from both
      cilium/cilium and cilium/hubble repositories and manually combine them.

      docker pull quay.io/cilium/release-tool:main
      alias release='docker run -it --rm -e GITHUB_TOKEN=$(gh auth token) quay.io/cilium/release-tool:main'
      release changelog --base v$PREVIOUS_RELEASE --head v$NEW_RELEASE --repo cilium/cilium --label-filter hubble-cli
      release changelog --base v$PREVIOUS_RELEASE --head main --repo cilium/hubble

- [ ] Modify `CHANGELOG.md` with the generated release notes:

      $EDITOR CHANGELOG.md
      ...
      git add CHANGELOG.md

- [ ] Push the prep branch and open a Pull Request against main branch.

       git commit -s -m "Prepare for v$NEW_RELEASE release"
       git push

     Get the pull request reviewed and merged.

- [ ] Update your local checkout:

      git checkout main
      git pull origin main

- [ ] Set the commit you want to tag:

      export COMMIT_SHA=<commit-sha-to-release>

     Usually this is the most recent commit on `main`, i.e.

      export COMMIT_SHA=$(git rev-parse origin/main)

- [ ] Create a tag:

      git tag -s -a "v$NEW_RELEASE" -m "v$NEW_RELEASE" $COMMIT_SHA

     Admire the tag you just created for 1 minute:

      git show "v$NEW_RELEASE"

     Then push the tag:

      git push origin "v$NEW_RELEASE"

- [ ] Ping [`hubble-maintainers` team] on Slack to get an approval to run
      [Image Release Build workflow].
- [ ] Wait for the [`Create a release` workflow] to finish.
- [ ] Find the release draft in the [Releases page]. Copy and paste release notes from
      CHANGELOG.md, and click on `Publish release` button.
- [ ] Update the [*Releases* section of the `README.md`] to point to the latest
      release.
- [ ] Update `stable.txt` in the main branch:

      git switch -c pr/$USER/update-stable-to-$NEW_RELEASE main
      echo v$NEW_RELEASE > stable.txt
      git add README.md stable.txt
      git commit -s -m "Update stable release to $NEW_RELEASE"
      git push origin pr/$USER/update-stable-to-$NEW_RELEASE

     and then open a pull request against the `main` branch, get it reviewed and merged.

[Cilium release tool]: https://github.com/cilium/release
[Image Release Build workflow]: https://github.com/cilium/hubble/actions/workflows/build-images-release.yaml
[`hubble-maintainers` team]: https://github.com/orgs/cilium/teams/hubble-maintainers
[Releases page]: https://github.com/cilium/hubble/releases
[Cilium Slack #general channel]: https://cilium.slack.com/archives/C1MATJ5U5
[*Releases* section of the `README.md`]: https://github.com/cilium/hubble/blob/main/README.md#releases
[`Create a release` workflow]: https://github.com/cilium/hubble/actions/workflows/release.yml
