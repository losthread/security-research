# My First Crackme Writeup

**Platform:** Windows x64  
**Difficulty:** Very Easy  
**Tool:** Ghidra

## 1. Finding the Main Function

After importing the executable into Ghidra, I identified the main function at `FUN_140001440`.

It reads the user's password using `fgets()` and passes it to `FUN_140001520` for validation.

## 2. Password Validation

The validation function contains:

```c
FUN_140001560(generated_password);
iVar1 = strcmp(password, generated_password);
return iVar1 == 0;
```

It generates the expected password and compares it with the user's input using `strcmp()`.

## 3. Decoding the Password

Inside `FUN_140001560`, I found:

```c
for (i = 0; i < 13; i++)
{
    output[i] = encoded[i] ^ 42;
}
```

The password is encoded using **XOR with 42 (`0x2A`)**.

Following the reference to `DAT_1400030e4` in the disassembly revealed these 13 bytes:

```text
62 19 46 46 45 75 69 58 4B 49 41 47 4F
```

Since XOR is reversible, applying the same key decodes the password:

```python
encoded = bytes.fromhex("62194646457569584b4941474f")
password = bytes(b ^ 42 for b in encoded).decode()
print(password)
```

Output:

```text
H3llo_Crackme
```

## 4. Flag

The main function prints the decoded password using:

```c
printf("Flag : CMO{%s}\n", decoded_password);
```

**Final Flag:**

```text
CMO{H3llo_Crackme}
```

## Conclusion

The crackme uses a simple XOR-based password obfuscation technique. By locating the validation function, examining the encoded bytes, and reversing the XOR operation, the password can be recovered without brute forcing or patching the binary.

### P.S

Challenge Link: https://crackmes.one/crackme/6ac65258cdee6e086a17364d
My Rating: 6/6
It was my first 'TRUE' R.E challenge, took around 3 hours and an existential crisis to solve.