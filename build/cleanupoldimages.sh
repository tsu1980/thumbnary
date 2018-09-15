#!/bin/bash
IMAGE="${1}"
DATE=`date -d "-${2} day" +%Y-%m-%d`
STAY_COUNT=${3}

IMAGE_LIST=$(gcloud container images list-tags ${IMAGE} --limit=999999 --sort-by=TIMESTAMP \
  --filter="timestamp.datetime < '${DATE}'" --format='get(digest)');
IMAGE_COUNT=$(echo "$IMAGE_LIST" | grep -v DIGEST | wc -l);

C=0
for digest in $IMAGE_LIST; do
  if [ $(( $IMAGE_COUNT - $C )) -ge $STAY_COUNT ]
  then
    (
      set -x
      gcloud container images delete -q --force-delete-tags "${IMAGE}@${digest}"
    )
    echo "container image(${IMAGE}@${digest}) deleted."
    let C=C+1
  else
    break
  fi
done

echo "Deleted $(( $C )) images that created before $DATE and not exceed remains count($STAY_COUNT) in ${IMAGE}."
