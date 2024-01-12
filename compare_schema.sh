#!/bin/bash

# Ask the user for the names of the output files
echo "Enter the name of the first output file:"
read file1
echo "Enter the name of the second output file:"
read file2

# Ask the user for the schema names
echo "Enter the schema name in the first output file:"
read schema1
echo "Enter the schema name in the second output file:"
read schema2

# Ask the user for the tablespace names
echo "Enter the tablespace name in the first output file:"
read tablespace1
echo "Enter the tablespace name in the second output file:"
read tablespace2

# Replace the schema names and tablespace names
sed "s/$schema1/SCHEMA_NAME/g; s/TABLESPACE $tablespace1/TABLESPACE TABLESPACE_NAME/g" $file1 > ${file1}_filtered
sed "s/$schema2/SCHEMA_NAME/g; s/TABLESPACE $tablespace2/TABLESPACE TABLESPACE_NAME/g" $file2 > ${file2}_filtered

# Compare the filtered files
diff ${file1}_filtered ${file2}_filtered > diff.txt

echo "Differences written to diff.txt."