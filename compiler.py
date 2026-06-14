import sys
import os
import re
import random
import math
import urllib.request
import json
import time

# =====================================================================
#   ██╗   ██╗ ██████╗ ██████╗ ████████╗███████╗██╗  ██╗
#   ██║   ██║██╔═══██╗██╔══██╗╚══██╔══╝██╔════╝╚██╗██╔╝
#   ██║   ██║██║   ██║██████╔╝   ██║   █████╗   ╚███╔╝ 
#   ╚██╗ ██╔╝██║   ██║██╔══██╗   ██║   ██╔══╝   ██╔██╗ 
#    ╚████╔╝ ╚██████╔╝██║  ██║   ██║   ███████╗██╔╝ ██╗
#     ╚═══╝   ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝
#
# DRONA LABS: VORTEX ENGINE - V1.0 ASCENDED CORE (PRO-DEV MASTER)
# ARCHITECTURE: TURING-COMPLETE, OS-INTEGRATED, WEB-ENABLED, JSON-NATIVE
# =====================================================================

# --- CUSTOM EXCEPTION HANDLING ---
class VortexSyntaxError(Exception):
    """Raised when the Lexer encounters an invalid character."""
    pass

class VortexRuntimeError(Exception):
    """Raised when the Engine encounters an execution fault (e.g., undefined variable)."""
    pass

# --- 1. THE LEXER (ENTERPRISE SYNTAX SCANNER) ---
# We break down the exact regex for every single language component.
TOKEN_REGEX = [
    ('COMMENT',   r'//.*'),                                      
    ('NUMBER',    r'\d+(\.\d+)?'),                               
    ('STRING',    r'".*?"|\'.*?\''),                             
    ('KEYWORD',   r'\b(fn|let|if|else|while|true|false|print|return|save|load|into|request)\b'),
    ('BUILTIN',   r'\b(random|sqrt|round|abs|pow|sin|cos|tan|length|type_of|str)\b'), # Added str()
    ('IDENTIFIER',r'[a-zA-Z_][a-zA-Z0-9_]*'),                    
    ('OPERATOR',  r'==|!=|>=|<=|>|<|=|\+|-|\*|/|&&|\|\||,'),     
    ('PUNCT',     r'[\[\]{}();:]'),                              
    ('SPACE',     r'\s+'),                                       
    ('UNKNOWN',   r'.'),                                         
]

def lexer(code):
    """
    Scans the raw .vx text file and converts it into actionable tokens.
    Removes whitespace and comments instantly to optimize execution speed.
    """
    tokens = []
    master_regex = '|'.join(f'(?P<{name}>{pattern})' for name, pattern in TOKEN_REGEX)
    
    for match in re.finditer(master_regex, code):
        kind = match.lastgroup
        value = match.group()
        
        if kind in ['SPACE', 'COMMENT']: 
            continue
            
        if kind == 'UNKNOWN': 
            raise VortexSyntaxError(f"Unexpected character '{value}' found in script.")
            
        tokens.append((kind, value))
        
    return tokens

# --- 2. MEMORY MANAGEMENT ARCHITECTURE ---
class VortexMemoryStack:
    """
    Acts as the physical RAM for the Vortex Engine.
    Handles scoping to ensure local function variables don't overwrite global variables.
    """
    def __init__(self, global_mem=None):
        self.memory = global_mem if global_mem is not None else {}

    def allocate(self, var_name, value):
        self.memory[var_name] = value

    def update(self, var_name, value):
        if var_name not in self.memory:
            raise VortexRuntimeError(f"Attempted to update undefined variable: '{var_name}'")
        self.memory[var_name] = value

    def fetch(self, var_name):
        if var_name not in self.memory:
            raise VortexRuntimeError(f"Memory Pointer Error: '{var_name}' does not exist.")
        return self.memory[var_name]

    def dump_core(self):
        """Returns the raw dictionary for the Python evaluation bridge."""
        return self.memory

# --- 3. THE TURING-COMPLETE INTERPRETER ---
class VortexEngine:
    def __init__(self, memory_stack=None, functions=None):
        self.stack = memory_stack if memory_stack is not None else VortexMemoryStack()
        self.functions = functions if functions is not None else {}
        self.return_value = None

    def format_value(self, val):
        """Secures data types before pushing them into the math parser."""
        if isinstance(val, str) and not (val.startswith('"') or val.startswith("'")):
            return f'"{val}"'
        return str(val)

    def evaluate(self, expression_tokens):
        """
        The absolute core of the language.
        Translates Vortex math, JSON pointers, and logic directly into native Python execution.
        """
        resolved_tokens = []
        i = 0
        
        # PASS 1: Intercept and Execute Custom Functions First
        while i < len(expression_tokens):
            kind, value = expression_tokens[i]
            
            if kind == 'IDENTIFIER' and i + 1 < len(expression_tokens) and expression_tokens[i+1][1] == '(':
                func_name = value
                i += 2
                paren_count = 1
                args_list = []
                current_arg = []
                
                # Extract the arguments passed to the function
                while i < len(expression_tokens) and paren_count > 0:
                    if expression_tokens[i][1] == '(': paren_count += 1
                    elif expression_tokens[i][1] == ')': paren_count -= 1
                    
                    if paren_count > 0:
                        if expression_tokens[i][1] == ',' and paren_count == 1:
                            args_list.append(current_arg)
                            current_arg = []
                        else:
                            current_arg.append(expression_tokens[i])
                    i += 1
                
                if current_arg: 
                    args_list.append(current_arg)
                    
                # Evaluate arguments before passing them into the local scope
                evaled_args = [self.evaluate(arg) for arg in args_list if arg]
                
                # Spin up isolated RAM and execute the function
                if func_name in self.functions:
                    params, body = self.functions[func_name]
                    local_stack = VortexMemoryStack(dict(self.stack.dump_core())) 
                    
                    for p, a in zip(params, evaled_args):
                        local_stack.allocate(p, a)
                        
                    inner = VortexEngine(local_stack, self.functions)
                    inner.execute(body)
                    
                    res = inner.return_value if inner.return_value is not None else 0
                    resolved_tokens.append(('NUMBER', str(res)))
                else:
                    raise VortexRuntimeError(f"Execution Error: Unknown function '{func_name}'")
                continue
                
            resolved_tokens.append((kind, value))
            i += 1

        # PASS 2: Construct the Evaluation String
        expr_str = ""
        i = 0
        while i < len(resolved_tokens):
            kind, value = resolved_tokens[i]
            
            # Direct Pointer to the RAM dictionary
            if kind == 'IDENTIFIER':
                self.stack.fetch(value) # Safety check
                expr_str += f'__mem__["{value}"]'
            
            # Extended Standard Library
            elif kind == 'BUILTIN':
                func_name = value
                i += 2 
                arg_tokens = []
                paren_count = 1
                
                # Gather everything inside the builtin's parentheses
                while i < len(resolved_tokens) and paren_count > 0:
                    if resolved_tokens[i][1] == '(': paren_count += 1
                    elif resolved_tokens[i][1] == ')': paren_count -= 1
                    if paren_count > 0: arg_tokens.append(resolved_tokens[i])
                    i += 1
                
                arg_val = self.evaluate(arg_tokens)
                
                # Mathematical & Utility Library
                if func_name == 'random': expr_str += str(random.uniform(0, float(arg_val)))
                elif func_name == 'sqrt': expr_str += str(math.sqrt(float(arg_val)))
                elif func_name == 'round': expr_str += str(round(float(arg_val)))
                elif func_name == 'abs': expr_str += str(abs(float(arg_val)))
                elif func_name == 'pow': expr_str += str(math.pow(float(arg_val), 2))
                elif func_name == 'sin': expr_str += str(math.sin(float(arg_val)))
                elif func_name == 'cos': expr_str += str(math.cos(float(arg_val)))
                elif func_name == 'tan': expr_str += str(math.tan(float(arg_val)))
                elif func_name == 'length': expr_str += str(len(arg_val))
                elif func_name == 'type_of': expr_str += f'"{type(arg_val).__name__}"'
                elif func_name == 'str': expr_str += f'"{str(arg_val)}"' # Explicit string cast
                continue 
            
            # Logic Gate Translation
            elif kind == 'KEYWORD' and value == 'true': expr_str += 'True'
            elif kind == 'KEYWORD' and value == 'false': expr_str += 'False'
            elif kind == 'OPERATOR' and value == '&&': expr_str += ' and '
            elif kind == 'OPERATOR' and value == '||': expr_str += ' or '
            else:
                expr_str += value
            i += 1

        try:
            # Execute Native Python with Injected Vortex RAM
            return eval(expr_str, {"__mem__": self.stack.dump_core()})
        except Exception as e:
            # THE V1.0 STRING COERCION FIX: JavaScript-style Type Conversion!
            # If Python fails because of string + number math, we catch it and auto-glue it.
            operator_values = [v for k, v in resolved_tokens if k == 'OPERATOR']
            if '+' in operator_values:
                parts = []
                current_part = []
                for tk, tv in resolved_tokens:
                    # Break the expression apart at the plus signs
                    if tk == 'OPERATOR' and tv == '+':
                        parts.append(current_part)
                        current_part = []
                    else:
                        current_part.append((tk, tv))
                if current_part: 
                    parts.append(current_part)
                
                # Recursively evaluate each piece and smash them together as strings
                glued_string = ""
                for part in parts:
                    chunk_val = self.evaluate(part)
                    if isinstance(chunk_val, str) and chunk_val.startswith(('"', "'")):
                        chunk_val = chunk_val[1:-1]
                    glued_string += str(chunk_val)
                return f'"{glued_string}"'
                
            # If it wasn't an addition error, return the raw syntax fallback
            return "".join([v for k, v in resolved_tokens])

    def get_block(self, tokens, start_idx):
        """Extracts code blocks safely, respecting nested curly brackets."""
        block_tokens = []
        brace_count = 1
        i = start_idx
        while i < len(tokens) and brace_count > 0:
            if tokens[i][1] == '{': brace_count += 1
            elif tokens[i][1] == '}': brace_count -= 1
            if brace_count > 0: block_tokens.append(tokens[i])
            i += 1
        return block_tokens, i

    def execute(self, tokens):
        """The Main Master Loop for Actionable Commands."""
        i = 0
        while i < len(tokens):
            kind, value = tokens[i]

            # --- SYSTEM I/O: FILE SAVING ---
            if kind == 'KEYWORD' and value == 'save':
                var_name = tokens[i+1][1]
                filename_tokens = []
                j = i + 3 
                while tokens[j][1] != ';':
                    filename_tokens.append(tokens[j])
                    j += 1
                filename = str(self.evaluate(filename_tokens)).strip('"').strip("'")
                
                with open(filename, 'w', encoding='utf-8') as f:
                    data = self.stack.fetch(var_name)
                    if isinstance(data, (dict, list)): 
                        f.write(json.dumps(data, indent=4))
                    else:
                        data = str(data)
                        if data.startswith('"') and data.endswith('"'): data = data[1:-1]
                        f.write(data)
                i = j
                continue

            # --- SYSTEM I/O: FILE LOADING ---
            elif kind == 'KEYWORD' and value == 'load':
                filename_tokens = []
                j = i + 1
                while tokens[j][1] != 'into':
                    filename_tokens.append(tokens[j])
                    j += 1
                filename = str(self.evaluate(filename_tokens)).strip('"').strip("'")
                var_name = tokens[j+1][1]
                
                if os.path.exists(filename):
                    with open(filename, 'r', encoding='utf-8') as f:
                        data = f.read()
                    try: 
                        self.stack.allocate(var_name, json.loads(data))
                    except:
                        try:
                            num = float(data)
                            if num.is_integer(): num = int(num)
                            self.stack.allocate(var_name, num)
                        except: 
                            self.stack.allocate(var_name, f'"{data}"')
                else:
                    self.stack.allocate(var_name, '"FILE_NOT_FOUND"')
                i = j + 2
                continue

            # --- WEB API: JSON FETCHING ---
            elif kind == 'KEYWORD' and value == 'request':
                url_tokens = []
                j = i + 1
                while tokens[j][1] != 'into':
                    url_tokens.append(tokens[j])
                    j += 1
                url = str(self.evaluate(url_tokens)).strip('"').strip("'")
                var_name = tokens[j+1][1]
                
                try:
                    req = urllib.request.Request(url, headers={'User-Agent': 'Vortex-Engine/1.0'})
                    response = urllib.request.urlopen(req)
                    web_data = response.read().decode('utf-8')
                    try:
                        self.stack.allocate(var_name, json.loads(web_data)) 
                    except:
                        self.stack.allocate(var_name, f'"{web_data}"')
                except Exception as e:
                    self.stack.allocate(var_name, f'"WEB_ERROR: {str(e)}"')
                i = j + 2
                continue

            # --- LANGUAGE: FUNCTION BLUEPRINTS ---
            elif kind == 'KEYWORD' and value == 'fn':
                func_name = tokens[i+1][1]
                i += 3 
                params = []
                while tokens[i][1] != ')':
                    if tokens[i][0] == 'IDENTIFIER': params.append(tokens[i][1])
                    i += 1
                i += 2 
                block_tokens, new_i = self.get_block(tokens, i)
                self.functions[func_name] = (params, block_tokens)
                i = new_i
                continue

            # --- LANGUAGE: FUNCTION RETURNS ---
            elif kind == 'KEYWORD' and value == 'return':
                i += 1
                expr_tokens = []
                while i < len(tokens) and tokens[i][1] != ';':
                    expr_tokens.append(tokens[i])
                    i += 1
                self.return_value = self.evaluate(expr_tokens)
                return 

            # --- MEMORY: VARIABLE ALLOCATION ---
            elif kind == 'KEYWORD' and value == 'let':
                var_name = tokens[i+1][1]
                i += 3 
                expr_tokens = []
                while i < len(tokens) and tokens[i][1] != ';':
                    expr_tokens.append(tokens[i])
                    i += 1
                self.stack.allocate(var_name, self.evaluate(expr_tokens))

            # --- MEMORY: VARIABLE RE-ASSIGNMENT ---
            elif kind == 'IDENTIFIER' and i + 1 < len(tokens) and tokens[i+1][1] == '=':
                var_name = value
                i += 2
                expr_tokens = []
                while i < len(tokens) and tokens[i][1] != ';':
                    expr_tokens.append(tokens[i])
                    i += 1
                self.stack.update(var_name, self.evaluate(expr_tokens))

            # --- UTILITY: TERMINAL OUTPUT ---
            elif kind == 'KEYWORD' and value == 'print':
                i += 2 
                expr_tokens = []
                paren_count = 1
                while i < len(tokens) and paren_count > 0:
                    if tokens[i][1] == '(': paren_count += 1
                    elif tokens[i][1] == ')': paren_count -= 1
                    if paren_count > 0: expr_tokens.append(tokens[i])
                    i += 1
                
                result = self.evaluate(expr_tokens)
                
                # Pretty Format for Terminal Output
                if isinstance(result, (dict, list)):
                    result = json.dumps(result, indent=4)
                elif isinstance(result, str) and result.startswith(('"', "'")) and result.endswith(('"', "'")):
                    result = result[1:-1]
                    
                print(f"VORTEX>> {result}")

            # --- LOGIC: IF / ELSE GATES ---
            elif kind == 'KEYWORD' and value == 'if':
                i += 2 
                cond_tokens = []
                paren_count = 1
                while i < len(tokens) and paren_count > 0:
                    if tokens[i][1] == '(': paren_count += 1
                    elif tokens[i][1] == ')': paren_count -= 1
                    if paren_count > 0: cond_tokens.append(tokens[i])
                    i += 1
                
                condition_result = self.evaluate(cond_tokens)
                block_tokens, new_i = self.get_block(tokens, i)
                i = new_i
                
                if condition_result:
                    inner = VortexEngine(self.stack, self.functions)
                    inner.execute(block_tokens)
                    if inner.return_value is not None:
                        self.return_value = inner.return_value
                        return
                
                if i < len(tokens) and tokens[i][1] == 'else':
                    i += 2 
                    else_tokens, new_i = self.get_block(tokens, i)
                    i = new_i
                    if not condition_result:
                        inner = VortexEngine(self.stack, self.functions)
                        inner.execute(else_tokens)
                        if inner.return_value is not None:
                            self.return_value = inner.return_value
                            return
                continue

            # --- RECURSION: WHILE LOOPS ---
            elif kind == 'KEYWORD' and value == 'while':
                i += 2
                cond_tokens = []
                paren_count = 1
                while i < len(tokens) and paren_count > 0:
                    if tokens[i][1] == '(': paren_count += 1
                    elif tokens[i][1] == ')': paren_count -= 1
                    if paren_count > 0: cond_tokens.append(tokens[i])
                    i += 1
                
                block_tokens, new_i = self.get_block(tokens, i)
                i = new_i
                
                loop_safety = 0
                while self.evaluate(cond_tokens) and loop_safety < 50000:
                    inner = VortexEngine(self.stack, self.functions)
                    inner.execute(block_tokens)
                    if inner.return_value is not None:
                        self.return_value = inner.return_value
                        return
                    loop_safety += 1
                    
                if loop_safety >= 50000:
                    print("[!] VORTEX WARN: Infinite loop threshold breached.")
                continue

            i += 1

# --- 4. THE BOOT SEQUENCE ---
def print_header():
    print("\n" + "="*65)
    print("  ██╗   ██╗ ██████╗ ██████╗ ████████╗███████╗██╗  ██╗")
    print("  ██║   ██║██╔═══██╗██╔══██╗╚══██╔══╝██╔════╝╚██╗██╔╝")
    print("  ██║   ██║██║   ██║██████╔╝   ██║   █████╗   ╚███╔╝ ")
    print("  ╚██╗ ██╔╝██║   ██║██╔══██╗   ██║   ██╔══╝   ██╔██╗ ")
    print("   ╚████╔╝ ╚██████╔╝██║  ██║   ██║   ███████╗██╔╝ ██╗")
    print("    ╚═══╝   ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝")
    print("         DRONA LABS: V1.0 ASCENDED CORE ONLINE")
    print("="*65 + "\n")

def main():
    if len(sys.argv) < 2: 
        print("Usage: python compiler.py <filename.vx>")
        return
        
    file_path = sys.argv[1]
    
    print_header()
    
    try:
        start_time = time.time()
        
        with open(file_path, 'r', encoding='utf-8') as file:
            code = file.read()
            
        tokens = lexer(code)
        engine = VortexEngine()
        engine.execute(tokens)
        
        end_time = time.time()
        print(f"\n[Vortex Runtime: {round((end_time - start_time) * 1000, 2)} ms]")
        
    except VortexSyntaxError as e:
        print(f"\n[X] SYNTAX PANIC: {str(e)}")
    except VortexRuntimeError as e:
        print(f"\n[X] RUNTIME PANIC: {str(e)}")
    except Exception as e:
        print(f"\n[X] CRITICAL CORE FAILURE: {str(e)}")

if __name__ == "__main__":
    main()