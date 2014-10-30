# Pre-requisites

You'll want to make sure you have [Consul](http://www.consul.io/intro/getting-started/install.html) installed on the node you're going to test with.  You don't need a full cluster to get going.

Next, let's install Daybook.  If you're comfortable with compiling and moving the binaries into place, great.  If not, you can use `go get` to pull it down, build it, and install it for you:

    go get github.com/tobz/daybook/daybook-pull
    go get github.com/tobz/daybook/daybook-push
    go get github.com/tobz/daybook/daybook-map

These are going to end up in $GOPATH/bin, so you'll want to make sure that's in your place.  You could do this:

    export PATH=$PATH:$GOPATH/bin

By default, Daybook looks for a bucket called "daybook", but you can create one with whatever name you want and specify that name in the Daybook configuration file -- you can find an example configuration file [here](https://raw.githubusercontent.com/tobz/daybook/master/daybook-pull/daybook.json.example); just fill in the bits you care about.  The quickest way to get started is to make a bucket called "daybook".

You'll also need a set of AWS credentials that can read and write to whatever your bucket is.  You'll need to put these in the Daybook configuration file, or in environment variables, like so:

    export AWS_ACCESS_KEY="..."
    export AWS_SECRET_KEY="..."

If you're on an instance with an IAM profile, make sure the profile has read/write permissions to the bucket, but other than that, Daybook will pick them up automatically - no configuration file or environment variables required.

By default, `daybook-pull` is going to install your service assets to `/tmp`.  We'll stick with that for now because we don't need to care about making another directory.

# Getting an asset into S3

First things first: we need something to pull down.  If you have a tarball handy, you can use it, but otherwise, here's some commands to create a simple one:

     mkdir /tmp/example-tarball
     cd /tmp/example-tarball
     echo "#\!/bin/bash\necho foo" > exec
     chmod +x exec
     echo "bar" > data
     tar cvzf ../example-tarball.tar.gz .
     cd /tmp
     rm -rf /tmp/example-tarball

Now you have a tarball at `/tmp/example-tarball.tar.gz`.  We're going to send that to S3 using `daybook-push` so that it puts it in the right spot for us.  We're going to call this the "example" service, and we'll say the version of this package is the current datetime.  So let's call `daybook-push`:

    daybook-push example $(date +%Y%m%d%H%M%S) /tmp/example-tarball.tar.gz

You should see some output like this:

    [toby:/tmp] daybook-push example $(date +%Y%m%d%H%M%S) /tmp/example-tarball.tar.gz
    2014/10/30 11:09:33 Sending /tmp/example-tarball.tar.gz as example/20141030110933...
    2014/10/30 11:09:33 All done!

Sweet.  We've got our first service asset in place.  If you want, you could go look in your S3 bucket and you're going to see a file called "example-20141030110933.tar.gz".  Now, let's add a mapping to Consul that puts the "example" service on our host.  We're going to create a mapping that catches all hosts, and progress from there.

There's a tool you haven't met yet, and that is `daybook-map`.  This is another helper tool to put mappings into Consul for you.  You could do this by hand, for sure, but again, just a helper tool.  Let's create our mapping to match all hosts (*) and for those hosts, give them the "example" service.

    daybook-map add "*" example

Make sure to use the double quotes, otherwise your shell will most likely glob all the files from your current directory and stick them in there, which is not what we want. :)  Now, to makesure our mapping is in Consul, run this:

    daybook-map list "*"

You should see output that looks like this:

    [toby:/tmp] daybook-map add "*" example
    2014/10/30 12:58:43 Adding services to '*'...
    2014/10/30 12:58:43 Services for '*': example
    2014/10/30 12:58:43 All done!

Boom.  It's in.  Now we're ready to pull down our assets!  This one is easy:

    daybook-pull

You should see output that looks like this:

    [toby:/tmp] daybook-pull
    2014/10/30 13:01:19 Getting services from register...
    2014/10/30 13:01:19 Got 1 service(s) from register!
    2014/10/30 13:01:19 Getting assets for service 'example'...
    2014/10/30 13:01:19 Pulling down example/20141030110933...
    2014/10/30 13:01:20 All done!

And if you look under /tmp, you should see something like this:

    [toby:/tmp] ls /tmp
    example example-tarball.tar.gz
    [toby:/tmp] ls /tmp/example
    20141030110933
    [toby:/tmp] ls /tmp/example/20141030110933
    data exec

Boom.  You just pulled down service assets to your host.
