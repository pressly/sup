#!/bin/bash
set -e

# Clone the repo for the first time.
if [[ ! -d /tmp/$IMAGE || ! -d /tmp/$IMAGE/.git ]]; then
	mkdir -p /tmp/$IMAGE
	git clone -b $BRANCH $REPO /tmp/$IMAGE
fi

cd /tmp/$IMAGE

# Fix permissions.
sudo chown -R $USER:$USER ./

# Clean up the repository.
git reset --hard && git clean -f -d

# Switch to master and update.
git checkout master && git pull

# Remove all branches except for master.
# (This prevents error when someone rebases his branch.)
for branch in $(git branch | grep -v master || :); do
	git branch -D $branch
done

# If we're deploying a custom branch, we need to pull it first.
if [ "$BRANCH" != "master" ]; then
	git checkout -b $BRANCH origin/$BRANCH
fi
