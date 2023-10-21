# https://github.com/helm/helm/issues/8036
if [ -f "${1}/Chart.yaml" ]; then
  yq --indent 0 '.dependencies | map(["helm", "repo", "add", .name, .repository] | join(" ")) | .[]' "${1}/Chart.yaml" | xargs -n1 bash -c;
fi
