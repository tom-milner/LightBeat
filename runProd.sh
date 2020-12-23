echo "Fetching latest from github..."

# Stash any changes
git stash

# Get latest from github
git reset --hard FETCH_HEAD


