#!/bin/bash
set -eu -o pipefail # -x
_wd=$(pwd); _path=$(dirname $0 | xargs -i readlink -f {})

command -v docker > /dev/null
command -v git > /dev/null
command -v yq > /dev/null

#### 1. setup
[ $# -eq 0 ] && { >&2 echo '!!! '"Argument {branch} is required!"; exit 1; }

git_branch=$1

app_name=$(yq .app_name project.yaml)
app_version=$(yq .app_version project.yaml)
image_name=$(yq .image_name project.yaml)

[[ "${app_name}${app_version}${image_name}" == *"null"* ]] &&
  { >&2 echo '!!! '"args are unset in project.yaml"; exit 1; }

image_tag=${git_branch}-${app_version}
image_tag=${DOCKER_Tag:-$image_tag}
image=$image_name:$image_tag
# build_time=$(date +'%FT%T.%N%:z')
build_time=$(date +'%FT%T%:z')
build_host=$(hostname)

# env variables
# GIT_Pull=$(printenv GIT_Pull || true)
GIT_Pull=${GIT_Pull:-"true"}
DOCKER_Pull=${DOCKER_Pull:-"true"}
DOCKER_Push=${DOCKER_Push:-"true"}
region=${region:-""}

[ -s .env ] && { 2>&1 echo "==> load .env"; . .env; }

#### 2. git
function on_exit() {
    git checkout dev #??? --force
}
trap on_exit EXIT

git checkout $git_branch

git_repository="$(git config --get remote.origin.url)"

git_branch="$(git rev-parse --abbrev-ref HEAD)" # current branch
git_commit_id=$(git rev-parse --verify HEAD) # git log --pretty=format:'%h' -n 1
git_commit_time=$(git log -1 --format="%at" | xargs -I{} date -d @{} +%FT%T%:z)
git_tree_state="clean"
uncommitted=$(git status --short)
unpushed=$(git diff origin/$git_branch..HEAD --name-status)
[[ ! -z "$unpushed" ]] && git_tree_state="unpushed"
[[ ! -z "$uncommitted" ]] && git_tree_state="uncommitted"

if [[ "$GIT_Pull" != "false" && ! -z "$uncommitted$unpushed" ]]; then
    >&2 echo '!!! '"git state is dirty"
    exit 1
fi

[[ "$GIT_Pull" != "false" ]] && git pull --no-edit

#### 3. pull image
[[ "$DOCKER_Pull" != "false" ]] && \
for base in $(awk '/^FROM/{print $2}' ${_path}/Dockerfile); do
    echo ">>> Pull image: $base"
    docker pull $base

    bn=$(echo $base | awk -F ":" '{print $1}')
    if [[ -z "$bn" ]]; then continue; fi
    docker images --filter "dangling=true" --quiet "$bn" | xargs -i docker rmi {}
done

#### 4. build image
echo "==> Building image: $image..."

mkdir -p cache.local proto

cat > cache.local/build.yaml << EOF
app_name: $app_name
app_version: $app_version
git_branch: $git_branch
git_commit_id: $git_commit_id
git_commit_time: $git_commit_time
git_tree_state: $git_tree_state

build_time: $build_time
image: $image
EOF

GO_ldflags="\
  -X main.build_time=$build_time \
  -X main.git_branch=$git_branch \
  -X main.git_commit_id=$git_commit_id \
  -X main.git_commit_time=$git_commit_time \
  -X main.git_tree_state=$git_tree_state \
  -X main.image_tag=$image_tag"

#???  -X main.git_repository=$git_repository -X main.image=$image"
#???  -X main.build_host=$build_host

docker build --no-cache --file ${_path}/Containerfile \
  --build-arg=BUILD_Time="$build_time" \
  --build-arg=region="$region" \
  --build-arg=APP_Name="$app_name" \
  --build-arg=APP_Version="$app_version" \
  --build-arg=GO_ldflags="$GO_ldflags" \
  --tag $image ./

#### 5. push image
[ "$DOCKER_Push" != "false" ] && docker push $image

#### 6. remove dangling images
docker image prune --force --filter label=stage=${app_name}_build &> /dev/null
# docker images --filter "dangling=true" --quiet $image | xargs -i docker rmi {}
for img in $(docker images -f "dangling=true" -f label=app=${app_name} --quiet); do
    >&2 echo "==> remove image: $img"
    docker rmi $img || true
done
