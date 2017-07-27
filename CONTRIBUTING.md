The Babelfish project is open source and we're happy to receive external
contributions. But, as in any licensed project, there is some important legal
information that you need to be aware of:

First, your contributions to each project will be licensed under the appropriate license
as found in the LICENSE file at the root directory of each project. 

Second, similar to what other open source projects do (i.e. the Linux kernel), we require
that you read and accept our [Developer Certificate of
Origin](https://developercertificate.org).

The way to easily communicate that you've read and accepted the DCO for every
contribution is to add a `Signed-off-by` line to your commits. With git, you can do
it by adding the `--signoff` parameter. The `Signed-off-by` line must have your
**real name and surname** and your email.

This is how a `Signed-off-by` line typically looks:

```git
Signed-off-by: John Foobar <jfoobar@someemail.com>
```

If you see something like `root <root@localhost>` instead, you need to configure
your identity in git. This is done with these commands:

```bash
$ git config --global user.name "John Foobar"
$ git config --global user.email jfoobar@someemail.com
```
