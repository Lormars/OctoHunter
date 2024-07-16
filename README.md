# OctoHunter

This is an in-dev multi-purpose bug hunting machine. The bug it hunts depend on the research/paper/vulnerability I am currently working on.

# Design Philosophy

This tool takes inspiration from the bbot by blacklanternsecurity to use a recursive way to find bugs. It has several modules, and can easily add more modules, and each module would interact with other modules such that the output (or middle product) of one module would be fed into other modules, and vice versa. This way, all you need is a simple list of subdomains, and then you can let the machine does its work.

Given this philosophy, there is also a preferred way of usage. This tool is not suit for doing quick scan of a small set of targets. Rather, it is better to run it continuously on a VPS on a large set of targets.

# Usage

There are three modes, and two of them are deprecating. The recommendated approach is to only use the dispatcher mode, which takes use of the design philosophy mentioned above. However, there are several pre-requisite to use this tool currently.

## Pre-requisite

1. Proxy: you need to have a list of proxies in order to use this tool.

to be continued...