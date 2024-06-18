## Features
1. Anchorfiles will be directly imported into a Dockerfile
1. Anchorfiles should be full-fledged Dockerfiles on their own, with the added benefit, that they can be imported into Dockerfiles
1. RUN instructions are not merged
1. All ARG and ENV instructions present in any of the Anchorfiles, must be declared in the Dockerfile, every subsequent declaration is removed;
If an ARG or ENV has a default value in one of the Anchorfiles, but not in the Dockerfile, the default value is added to the Dockerfile and is not overriden further;
1. If there are LABELs with the same key, an error is thrown, or if there are LABELs with the same value, a Warning is shown, but compilation continues
1. Each Anchorfile is imported only once, and only the names are checked, so if there are 2 Anchorfiles with the same name, but different paths, only the first encountered file is checked;
If an Anchorfile has no special name (i.e., is just "Anchorfile"), the name of the directory is included in the name (e.g. "myservice/Anchorfile")
1. There can be only one FROM, and that is the one in the Dockerfile. Every FROM is removed from an Anchorfile; If there are multiple FROMs encountered in any of the files
1. The path in every COPY and ADD instruction is adjusted, so that the paths is still valid, even though the work directory has changed
1. If there a port is present in multiple PORTs, no matter if internal or external, an error is thrown
1. If there are multiple CMDs, it will show a warning 
1. Because WORKDIRs are relative to each other, if there are multiple WORKDIRs a warning is shown
1. Any MAINTAINERs instruction will be removed, and a warning will be logged
### Stretches
1. Anchorfiles can be fetched from the hub
## Internals
### Operation Order
1. First the Lexical and Syntax Analysis is done on the main Dockerfile (Tokenizing and building the AST)
1. Then the same the steps are applied for the each Anchorfile, and the Nop Node of each Anchorfile replaces the Comment Node, from where they were anchored
1. Though before an Anchorfile is processed, it is checked that if it has been processed before, and if it has, it is skipped
1. Go through the AST, bottom to top, and verify that no ports are declared multiple times, paths are valid, etc. 
    - Bottom to top, because this way, if an Anchorfile defines an ARG with a default value and Dockerfile does not, 
    the last encountered ARG with a default value can be used to swap out the ARG in the Dockerfile, and this way, the first default value is used throughout as intended
1. Lastly it is compiled into the final Dockerfile, including optional comments from the original files, and also comments marking from which file the layer is from

## Instruction specific details
- Each Instruction is represented as a Node in the syntax tree, where each node has these properties:
    - token: instruction name, or other special symbol that has a meaning ("--", "#", etc.); is empty when the node represents a command
    - value: e.g. the entirety of a command
    - children: node needs to have either a value or child nodes, usually shouldn't be empty nor should it have both 
- Each Instruction has acts as a Top Node of the layer
- Special internal keywords are:
    - "SRC" - indicating a source file path, from the local machine, so that it can be identified and used to adjust the path
    - "DEST" - indicating a destination file path in the container, is not modified
    - "[]" - indicating an array of items, and the contents of children will be surrounded by brackets and joined with a ","

### ADD
- Child nodes are either options (they start with "--"), or paths (SRC + DEST)
- Option nodes either have a value (e.g. "--chown", "--link"), or have a child node for key value pairs (e.g. "--keep-git-dir=true")
- The output will produce the command in this format: `ARG [OPTIONS] "SRC" "DEST"`
### ARG 
- Has a single child node, where the token is the key, and value is the default value of the arg
### CMD 
- No special instruction, the output will be the same as the input
### COPY
- Same as ADD
### ENTRYPOINT
- Same as CMD
### ENV 
- Same as LABEL
### EXPOSE
- Doesn't have a child node and only includes the value, which is "port/protocol" (e.g. 80/udp)
### FROM
- Can have the following children:
    - "--", which then has another child for key-value-pair
    - the image name, including the tag and digest
    - alias, which has token AS and then the value is the alias name
### HEALTHCHECK
    - Either has value NONE, or has multiple children
    - If has children, first they are key-value-pairs, and then final child is a CMD Instruction
### LABEL 
- Similar to ARG, but with the difference where it can have multiple key-value-pairs in a single layer
### MAINTAINER
- Same as CMD
### ONBUILD
- Has only a single child, and that will be any other instruction according to it's rules
### RUN
- Has key-value-pairs as options (WARNING, "=" can appear in the value, e.g. "--mount=type=secret", where the "mount" is the key, and "type=secret" is the value)
- After options, the entire command is the last child
### SHELL
- Has a single child, which will be [] and the value will be string array of extracted values
### STOPSIGNAL
### USER
### VOLUME
### WORKDIR


## Other details
- The Top Node and Comment Node need to be interchangable

