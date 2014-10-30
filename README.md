Daybook
=======

Opinionated service deployment management.

![daybook](http://i.imgur.com/WNo8CMU.png)

What the heck does it do?
=======

Daybook provides the basis for a pull-based deployment model that doesn't break the mindshare bank.  Simple tools, simple primitives.  

The premise is, wait for it, simple: you put all of your service assets -- these are your tarballs of service code, JARs, data files, whatever -- in a dedicated bucket on S3.   Next, you set up a datacenter-wide key/value store (Consul) to store the mappings of what services should be on what hosts.  These mappings are flexible, and can include wildcards.  Specificity wins.  You structure the name of your service assets in a way that the name of a service, or services, in your mapping data can be used to find the assets in your bucket.  Voila, now you're ready to pull down some assets to different hosts without adding too much entropy to your infrastructure, or the universe.

How do I use it?
=======

Check out the [quick start guide](docs/quick-start-guide.md) for hitting the ground running.

Tell me more about the specifics
=======

`daybook-pull` makes some assumptions.  Firstly, your assets are structured in a particular fashion.  This is roughly enforced by `daybook-push`, if you use it, but since it's just S3, there's no reason you can't upload assets on your own.  

Your assets need to meet the following requirements:

- it has to be a tarball (gzip compressed tar archive, with the ".tar.gz" extension; this is hard-coded)
- the filename must be in the form of: service_name-version.tar.gz
- the service name can be any alphanumeric characters, including underscores.  hyphens aren't allowed because they're the name/version delimiter
- the version can also be alphanumeric characters, including underscores
- the extension, as mentioned above, needs to be ".tar.gz"

Secondly, it assumes you want all versions of a given service on disk.  Service discovery, and thus which version to use, is an entirely separate problem that Daybook doesn't try to solve.  Disk space is cheap, and it's simpler to assume we want a synchronized universe of service code than to pick and choose.  There may or may not be a future improvement to limit scope or try and detect what we already have... but it's not on the immediate horizon.  Sorry, people with a million JAR files that are 90MB a piece.

When you do a pull, `daybook-pull` will load the configuration, connect to Consul and look for patterns that match the hostname it has.  The hostname is either what's in the configuration file, or if that's empty, what it is able to get from the OS.  If it can't do that, it will whine and exit.

For all the patterns it finds that could match the hostname given, it figures out which one is most specific.  The patterns can use asterisks for a wildcard.  They aren't full regular expressions, because trying to determine specificity from a regular expression is non-trivial.  Wildcards work 99% of the time, and the longest matching pattern is implicitly the most specific pattern.  Simple.

In Consul, using its key/value store, these patterns look something like this:

    /v1/kv/daybook/hosts/web-prod-* -> service_1,service_2,service_3
    /v1/kv/daybook/hosts/web-prod-oneoff-* -> service_1,service_4

So, you can see that for a hostname of "web-prod-oneoff-001", both of those keys would match.  Due to the length, "web-prod-oneoff-*" would be the most specific match, and so we'd try and download all of the service assets for "service_1" and "service_4".  Conversely, if you had mappings that looked like this:

    /v1/kv/daybook/hosts/web-prod-* -> service_1,service_2,service_3
    /v1/kv/daybook/hosts/web-prod-oneoff-* -> service_1,service_4
    /v1/kv/daybook/hosts/web-prod-oneoff-001 -> service_5,service_6
    
Again, all of these patterns would match the same hostname, but we have a full match, which implicitly means it's the longest, and so we'd try and download all of the service assets for "service_5" and "service_6".  Also, as you can see, tThe services are specified as a comma-delimited list, which keeps things... you got it, simple.

For each service, we query the S3 bucket for all objects that have a perfix of the service name.  For all the objects we get back, we make sure they conform -- ".tar.gz" extension, matches the service_name-version naming scheme, etc.  If it looks good, we decompress it and extract it on the fly.  Based on the configuration, we use the specified installation directory, and within that, make two subdirectories: one for the service, and another under that for the version.  So, for a serivce asset called

    test_service-123.tar.gz
    
it would be extracted into

    <install directory>/test_service/123/
    
Thus, you'll want your archiving process to reflect that and not bother with encapsulating all of the assets within a single folder before tarballing.
