#!/bin/bash

unameOut="$(uname -s)"
case "${unameOut}" in
    Linux*)     machine=Linux;;
    Darwin*)    machine=Mac;;
    CYGWIN*)    machine=Cygwin;;
    MINGW*)     machine=MinGw;;
    *)          machine="UNKNOWN:${unameOut}"
esac

READLINK=readlink
if [ "$machine" = "Mac" ];then
  READLINK=greadlink
fi


CURRENT_DIR=`dirname $($READLINK -f $0)`
ROOT_DIR=${CURRENT_DIR%%/build}

cd $ROOT_DIR
BUILD_ROOT_DIR=$ROOT_DIR/build/

HEP_TYPE=$1
PUBLISH_FLAG=$2


get_version_vars() {
  ROOT=$1
  local git=(git --work-tree "${ROOT}")

  if [[ -n ${GIT_COMMIT-} ]] || GIT_COMMIT=$("${git[@]}" rev-parse "HEAD^{commit}" 2>/dev/null); then
    if [[ -z ${GIT_TREE_STATE-} ]]; then
      # Check if the tree is dirty.  default to dirty
      if git_status=$("${git[@]}" status --porcelain 2>/dev/null) && [[ -z ${git_status} ]]; then
        GIT_TREE_STATE="clean"
      else
        GIT_TREE_STATE="dirty"
      fi
    fi

    # Use git describe to find the version based on annotated tags.
    if [[ -n ${GIT_VERSION-} ]] || GIT_VERSION=$("${git[@]}" describe --tags --abbrev=14 "${GIT_COMMIT}^{commit}" 2>/dev/null); then
      # This translates the "git describe" to an actual semver.org
      # compatible semantic version that looks something like this:
      #   v1.1.0-alpha.0.6+84c76d1142ea4d
      #
      # TODO: We continue calling this "git version" because so many
      # downstream consumers are expecting it there.
      DASHES_IN_VERSION=$(echo "${GIT_VERSION}" | sed "s/[^-]//g")
      if [[ "${DASHES_IN_VERSION}" == "---" ]] ; then
        # We have distance to subversion (v1.1.0-subversion-1-gCommitHash)
        GIT_VERSION=$(echo "${GIT_VERSION}" | sed "s/-\([0-9]\{1,\}\)-g\([0-9a-f]\{14\}\)$/.\1\+\2/")
      elif [[ "${DASHES_IN_VERSION}" == "--" ]] ; then
        # We have distance to base tag (v1.1.0-1-gCommitHash)
        GIT_VERSION=$(echo "${GIT_VERSION}" | sed "s/-g\([0-9a-f]\{14\}\)$/+\1/")
      fi
      if [[ "${GIT_TREE_STATE}" == "dirty" ]]; then
        # git describe --dirty only considers changes to existing files, but
        # that is problematic since new untracked .go files affect the build,
        # so use our idea of "dirty" from git status instead.
        GIT_VERSION+="-dirty"
      fi

      GIT_VERSION_SHORT=${GIT_VERSION%%+*}
      # Try to match the "git describe" output to a regex to try to extract
      # the "major" and "minor" versions and whether this is the exact tagged
      # version or whether the tree is between two tagged versions.
      if [[ "${GIT_VERSION}" =~ ^v([0-9]+)\.([0-9]+)(\.[0-9]+)?([-].*)?$ ]]; then
        GIT_MAJOR=${BASH_REMATCH[1]}
        GIT_MINOR=${BASH_REMATCH[2]}
        if [[ -n "${BASH_REMATCH[4]}" ]]; then
          GIT_MINOR+="+"
        fi
      fi
    fi
  fi
}

build_image() {
  type=$1

  dockerfile=$BUILD_ROOT_DIR/image/Dockerfile
  build_context=$ROOT_DIR

  image_base=registry.cn-beijing.aliyuncs.com/tinet-hub/heplify-server
  image_tag=${image_base}

  get_version_vars $ROOT_DIR
  ARG_VERSION=$GIT_VERSION
  DOCKER_VERSION=$GIT_VERSION_SHORT
  BUILD_TIME="ARG_BUILDTIME=$(date '+%Y-%m-%d+%H:%M:%S')"

  local -r latest_tag="${image_tag}:latest"
  local -r gitversiontag="${image_tag}:${CTI_CLOUD_DOCKER_VERSION}"


  docker build --build-arg "$BUILD_TIME" --build-arg "TYPE=$type"  \
    --build-arg "ARG_VERSION=$ARG_VERSION"  --rm -t "${gitversiontag}" -f "$dockerfile" "${build_context}"

  local -ra tag_cmd=(docker tag $gitversiontag $latest_tag)
  docker_tag_output=$("${tag_cmd[@]}")

  # return git version image name
  echo $gitversiontag

  publish=$2
  if [ "$publish" = "TRUE" ];then
    docker push $gitversiontag
    docker push $latest_tag
    docker rmi $gitversiontag
    docker rmi $latest_tag
    yes | docker image prune
  fi
}

function usage()
{
  echo "---------------------------------usage------------------"
  echo "-------./make-build-image.sh [ type ] ------------------"
  echo "-------type: hepsrv-------------------------------------"
}

case $HEP_TYPE in
hepsrv)
  printf "**********will build image $HEP_TYPE\r\n"
  ;;
*)
  usage
  exit 1
  ;;
esac

build_image $HEP_TYPE $PUBLISH_FLAG
