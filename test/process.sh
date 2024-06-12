#!/bin/bash

subfinder -dL takeover_domains -o subdomains
shuffledns -l takeover_domains -w $HOME/wordlist/SecLists/Discovery/DNS/bug-bounty-program-subdomains-trickest-inventory.txt -r $HOME/wordlist/resolvers.txt -mode bruteforce -o shuffle
cat subdomains shuffle | sort | uniq > all_subdomains

dnsx -l all_subdomains -nc -cname -re -o cnames_raw

cat cnames_raw | grep -iv "shop.spacex.com" > cnames