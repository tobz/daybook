daybook
=======

opinionated service deployment management

![daybook](http://i.imgur.com/WNo8CMU.png)

whats it do
=======

daybook provides the basis for a pull-based service deployment model. 

first, you store mappings in Consul that specify a pattern -- match a specific host, or use wildcards to match many -- and the names of services that belong on the hosts that match.

next, put your service assets into S3.  you'll need a dedicated bucket.  these assets are deployable versions of your service, in tar.gz format.  that's all daybook supports right now.  oh, and the name of the object must be "service_name-version.tar.gz".  there shouldn't be more than one hyphen, use underscores if you need to separate things in the service name.  i told you this was opinionated, right?

once that's in place, you need to make sure the local Consul agent is running.  don't use Consul in that way?  too bad, that's the idiomatic way.  you also need to specify an AWS credential pair in your environment variables in the typical fashion (AWS_ACCESS_KEY/AWS_SECRET_KEY or AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY) or give your instance, if on AWS, an IAM profile for instance credentials.  these credentials need to be able to read out of your assets bucket.

when all of that is done, you run daybook-pull, no arguments.  it takes the hostname where you're running it, queries Consul for all Daybook host mappings that match, and uses the most specific one.  it takes whatever services are specified for that mapping, and downloads all versions to the machine.

it's simple, kind of constrained, but it works, and it's fucking easy.  you like easy, don't you?

how do i use it
========

something like:

    go get github.com/tobz/daybook/daybook-pull
    $GOPATH/bin/daybook-pull
  
