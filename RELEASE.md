# RELEASE

Release process and checklist for `hubble`. 

## Prep the variables

These variables will be used in the commands throughout the README to allow
copy-pasting.

### Release hash

Identify which commit will serve as the release point.

New major and minor version with `.0` patch have to stem from the master
branch, while new patch releases have to stem from their respective minor
branches.

    export RELEASE_HASH=<commit hash, i.e. 37c8023>

### Version

If releasing a new version 5.4.0 with the latest release being 5.3.8, for
example, they will look as follows:

    export MAJOR=5
    export MINOR=4
    export PATCH=0
    export LAST_RELEASE=5.3.8

## Create release branch

If `.0` patch version is being created, a new `major.minor` branch has to be
made first. That branch will serve for tagging all releases, as well as
pointing to the latest patch release.

    git checkout -b v$MAJOR.$MINOR $RELEASE_HASH

NOTE: Do not directly commit to this branch. Follow the process and open a Pull
Request from the prep branch.

## Create release prep branch

This branch will be used to prepare all the necessary things to get ready for
release.

    git checkout -b v$MAJOR.$MINOR.$PATCH-prep

## Prepare the release notes

Using https://github.com/cilium/release, prepare the release notes between the
last minor version (latest patch) and current.

    ./release --repo cilium/hubble --base v$LAST_RELEASE --head v$MAJOR.$MINOR
    **Bugfixes:**
    * api: fix potential panic in endpoint's EqualsByID (#199, @Rolinh)

    **Other Changes:**
    * actions: Trigger on release branches (#233, @michi-covalent)
    * Add changelog (#203, @glibsm)
    * Adjust to moved PolicyMatchType location (#222, @tgraf)
    * api: Small fixes to the protoc invocations in Makefile (#206, @gandro)
    ... etc ...

Modify `CHANGELOG.md` with the generated release notes. Keep them handy, as
the same notes will be used in the github release.

    $EDITOR CHANGELOG.md
    ...
    git add CHANGELOG.md
    git commit -s -m "Modify changelog for $MAJOR.$MINOR.$PATCH release"

## Modify the version constant in the Makefile to match the new release

Usually this only consists of dropping the `-dev` suffix from the string.

    VERSION="$MAJOR.$MINOR.$PATCH"

Commit and push the changes to the prep branch

    git add Makefile
    git commit -s -m "Modify version to $MAJOR.$MINOR.$PATCH"

## Push the prep branch and open a Pull Request

The pull request has to be `v$MAJOR.$MINOR.$PATCH-prep -> v$MAJOR.$MINOR`

Once the pull request is approved and merged, a tag can be created.

## Modify the version constant on the master branch, if needed

After branching out from the tree for release, the version need to be updated
to reflect the next planned release, i.e.

    VERSION="$MAJOR.<$MINOR+1>.0-dev"

## Update the changelog in the master branch

Once the release PR has been merged, the changelog in the master branch needs to
be updated as well. Make sure to copy the generated release notes to the
changelog in the master.

## Update releases table in the readme file

The README file contains a section which lists all currently supported releases
in a table. The version in this table needs to be updated to match the new
release.

## Create a GitHub release

It is better to have github create the final release tag, rather than pushing
it through git. Pushing through git will auto-create an empty release and
notify all the users before you have a chance to include the list of changes,
or any other metadata.

https://github.com/cilium/hubble/releases/new

    Tag version:            v$MAJOR.$MINOR.$PATCH
    Target:                 v$MAJOR.$MINOR
    Release title:          same as tag version
    Describe this release:  Paste the earlier generated release notes

    Check the "This is a pre-release" box if `-rc*` or `0.x.x` release

## Finally, upload release tarballs to the GitHub release

Generate the release tarballs using `contrib/scripts/release.sh` script:

    make release

This will generate tarballs and associated checksum files in the `release`
directory. Make sure to upload these tarball and checksum to the GitHub release
page.

## Update the README.md

Update the *Releases* section of the `README.md` to point to the latest
GitHub release.

## (OPTIONAL) Update `stable.txt` in the master branch

Hubble's installation instruction in the Cilium documentation uses the version specified in
`stable.txt` in the master branch. There are a couple of things to consider when deciding
whether to update `stable.txt`. Let's say `stable.txt` is currently pointing to `v0.6.1`:

- If this is a minor or patch release relative to the current stable version (i.e. `v0.7.0`
  or `v0.6.2`), update `stable.txt` so that people start picking up the new features / bug
  fixes included in this release.
- If this is a patch release of a previous version (e.g. `v0.5.2`), don't update
  `stable.txt`.
- If this is a major release (e.g. `v1.0.0`), the installation instructions in older Cilium
  documentation versions need to be updated to point to a compatible version of Hubble. Then,
  ensure the version specified in `stable.txt` is compatible with the current stable Cilium
  version.

To update `stable.txt`, do:

    git checkout -b update-stable-txt master
    echo v$MAJOR.$MINOR.$PATCH > stable.txt
    git add stable.txt
    git commit -as -m "Point stable.txt to $MAJOR.$MINOR.$PATCH"
    git push

and then open a pull request against the master branch.
