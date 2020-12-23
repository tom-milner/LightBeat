echo "Fetching latest from github..."

# Stash any changes
git stash

# Get latest from github
git pull origin master

# Build Project
go build 

# Run!!
./LightBeatGateway


