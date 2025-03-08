# **enuma-elish-interpreter**

## **Introduction**

The **Enuma-Elish Interpreter** is a DSL (Domain-Specific Language) executor that allows executing **HTTP requests** and
**Python scripts** within a structured task execution pipeline. The interpreter reads `.ea` files, processes tasks, and
executes them based on the defined workflow.

## **Installation**

### **Prerequisites**

Ensure that you have the following installed on your system:

- **Go 1.18+** (for building the interpreter)
- **Docker** (for executing Python tasks in containers)
- **Bash** (for running installation scripts on Unix-based systems)

### **Building the Interpreter**

Clone the repository and navigate into the project directory:

```sh
git clone https://github.com/your-repo/enuma-elish-interpreter.git
cd enuma-elish-interpreter
```

### **Installing as a Global Command**

To install the interpreter globally on your system, run:

```sh
./zz_scripts/install-ea.sh
```

This script will:

- Copy the `ea-interpreter` binary to `/usr/local/bin/` (Linux/macOS)
- Make it accessible as a global command: `enumago`

To uninstall:

```sh
./zz_scripts/uninstall-ea.sh
```

## **Usage**

Once installed, you can execute `.ea` files using:

```sh
enumago path/to/file.ea
```

Example:

```sh
enumago samples/py-pick-scriptPath.ea
```

### **Running Without Installation**

If you don't want to install it globally, you can run it directly:

```sh
go run main.go path/to/file.ea
```

## **Troubleshooting**

- **Command Not Found**: Run `chmod +x zz_scripts/install-ea.sh` before running the installation script.
- **Permission Errors**: Use `sudo ./zz_scripts/install-ea.sh` to install globally.
- **Docker Issues**: Ensure Docker is running and accessible if you want to use Python tasks(`docker ps` should work).
