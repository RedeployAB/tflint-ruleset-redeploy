# terraform_output_order

## What does this rule do?

This rule enforces that output blocks are declared in alphabetical order based
on their names. It examines the order in which the output blocks appear in the
file and raises an error if they are not sorted alphabetically.

## Why is this important?

Alphabetical ordering of outputs helps maintain consistency and makes it easier
for team members to locate a specific output.

## How to fix issues

Reorder the output blocks in your configuration so that their names are in
alphabetical order.
