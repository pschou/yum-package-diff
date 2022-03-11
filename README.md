# Yum Package Diff

This shim takes two yum primary.xml.gz files, in the order old then new for
determining the files which have shown up or changed.  The intended purpose of this
shim is to be able to generate a file list for downloading.

# Example usage:
```bash
./yum-package-diff -new NewPrimary.xml.gz -old OldPrimary.xml -showAdded -output filelist.txt
```

and the output looks like:
```
$ ./yum-package-diff -new NewPrimary.xml.gz -old OldPrimary.xml -showAdded -output filelist.txt
2022/03/11 10:21:00 Reading in file NewPrimary.xml.gz
Using gz decoder
2022/03/11 10:21:02 Reading in file OldPrimary.xml
2022/03/11 10:21:04 doing matchups

$ cat filelist.txt
# Yum-diff matchup, version: 0.1.20220311.0830
# new: NewPrimary.xml.gz old: OldPrimary.xml
{sha256}35f6b7ceecb3b66d41991358113ae019dbabbac21509afbe770c06d6999d75c7 1818404 7/os/x86_64/Packages/389-ds-base-1.3.10.2-6.el7.x86_64.rpm
{sha256}e595924b51a69153c2148f0f4b3fc2c31a1ad3114a6784687520673740e4f54a 289524 7/os/x86_64/Packages/389-ds-base-devel-1.3.10.2-6.el7.x86_64.rpm
```


# Usage help:
```bash
$ ./yum-package-diff -h
Yum Package Diff,  Version: 0.1.20220310.1123

Usage: ./yum-package-diff [options...]

  -new string
        Package list for comparison (default "NewPrimary.xml.gz")
  -old string
        Package list for comparison (default "OldPrimary.xml.gz")
  -output string
        Output for comparison result (default "-")
  -repo string
        Repo path to use in file list (default "/7/os/x86_64")
  -showAdded
        Display packages only in the new list
  -showCommon
        Display packages in both the new and old lists
  -showRemoved
        Display packages only in the old list
```




