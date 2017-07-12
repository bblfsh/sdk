The Babelfish project is open source and we're happy to receive external
contributions. But, as in any licensed project, there is some important legal
information that you need to be aware of.

First, most of the components of the Babelfish project are licensed under the GPU
v3 or the Apache License 2.0. So, in order to contribute, you must accept to license
your contribution using either one of those licenses, depending on the subproject
(every project should have a LICENSE file with the specific license used).

Also, we require that you read and accept our [Developer Certificate
of Origin](https://developercertificate.org) which is exactly the same one that
the Linux kernel uses. You can find a copy in this repository in the file `DCO` in
the project root.  Please read it, but it basically says that you, for the best of
your knowledge, know that the code you're contributing doesn't have licensing
problems.

The way to easily communicate that you've read and accepted it for a specific
contribution is to add a `Signed-off-by` line to your commits. With git you can do
it by just adding the `--signoff` parameter. The `Signed-off-by` line must have
your **real name and surname** and your email.

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
