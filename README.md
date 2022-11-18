# Scrubber

Scrubber will anonymise files and folder by

- renaming files to file_(integer).extension
- renaming folders to dir_(integer)
- overwrite files with random data to the same size as the original had

Note that this tool is not made for securely wiping hardrives, specialised 
tools should be used for that purpose.

This tool is useful for creating test data from real data that can be used for
load testing, performance testing etc.

## Warning ##

If you haven't already realised it, this machine kills files and folders. 
You have been warned.

## Usage

```
usage: scrubber [-f | -v | -d ] ./path/to/directory
  -d	dry run, displays the operations that would be performed without actually running them
  -f	force, don't ask for confirmation, assume yes
  -v	verbose, show actions for all files
```

## Todo

- Faster random byte generation
