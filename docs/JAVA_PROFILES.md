# Java Profiles: Official Termux Stable Architecture & Differences (October 2026)

This document provides a comprehensive guide on the Java profile system in `termux-android-development` (`tadev`). It explains why profiles exist, details what each official OpenJDK version provides, and compares dual-JDK combinations against single-JDK solos.

Online Documentation: [https://tadev.vercel.app/docs/java-profiles](https://tadev.vercel.app/docs/java-profiles)

---

## 1. Exclusively Official Termux Stable Packages

`tadev` is built on a philosophy of maximum reliability and stability:

- **100% Official Termux Repository**: Every supported profile relies exclusively on packages from the official Termux stable repository (`openjdk-25`, `openjdk-21`, `openjdk-17`).
- **No Third-Party Repositories (Zero TUR)**: Unlike unstable setups, `tadev` does not require the Termux User Repository (`tur-repo`) or community-maintained third-party package sources.
- **Zero Termux:API Dependency**: You do **not** need to install the Termux:API companion APK (`com.termux.api`) or the `termux-api` package. Pressing `[Tab]` in the Java selection menu launches your device browser using native Android Activity Manager intents (`am start --user 0`) and base Termux bootstrap utilities (`termux-open-url`).

---

## 2. Termux Android Architecture & The Java Dilemma

Android app compilation inside Termux involves two distinct execution layers:

### Layer 1: Host Runtime (`JAVA_HOME` & `PATH`)
- Used by your interactive shell, terminal utilities, CLI scripts, and the Gradle daemon process launcher.
- Benefits from the latest performance improvements, compact object headers, and JIT optimizations available in newer OpenJDK releases.

### Layer 2: Gradle Compilation Toolchain (`org.gradle.java.installations.paths`)
- Used by the Android Gradle Plugin (AGP) and Kotlin compiler daemon (`kotlinc`) to compile Java/Kotlin source code into `.class` files, execute annotation processors (KSP, Dagger/Hilt), and run `d8`/`r8` to produce Dalvik Executable (DEX) bytecode.
- Official AGP 8.x and AGP 9.x releases target Java 17 and Java 21 bytecode. Compiling directly against unsupported cutting-edge host runtimes can trigger `Unsupported class file major version 69` errors or plugin crashes.

To solve this dilemma cleanly, `tadev` provides both **Dual-JDK Combinations** and **Single-JDK Solos** from official packages.

---

## 3. What Each Official Java Version Does (2026 Landscape)

### OpenJDK 25 (LTS - Released September 2025)
- **Role**: Modern default LTS runtime for Termux.
- **Key Capabilities**:
  - **Compact Object Headers (Project Lilliput)**: Drastically reduces JVM object header memory overhead, saving significant RAM on Android devices.
  - **Generational Shenandoah GC**: Ultra-low pause time garbage collector optimized for responsive interactive builds.
  - **Project Leyden AOT Optimizations**: Speeds up JVM startup and warm-up cycles for CLI tools.
  - **Flexible Constructor Bodies (JEP 513)**: Permits statements before `super()` or `this()` calls.
  - **Scoped Values**: Lightweight, secure alternative to thread-local variables.
  - **Native PEM Encoding**: Standardized parsing and generation of PEM security certificates.

### OpenJDK 21 (LTS - Released September 2023)
- **Role**: The gold standard for modern Android development. 100% supported by all AGP 8.x and 9.x releases.
- **Key Capabilities**:
  - **Virtual Threads (Project Loom)**: High-throughput lightweight concurrency.
  - **Record Patterns & Pattern Matching for Switch**: Expressive data deconstruction.
  - **Sequenced Collections**: Defined encounters for ordered collections (`getFirst()`, `getLast()`).
  - **Generational ZGC**: Scalable low-latency memory management.
  - **Key Encapsulation Mechanism (KEM) API**: Modern cryptographic primitives.

### OpenJDK 17 (LTS - Released September 2021)
- **Role**: Maximum legacy compatibility baseline. Minimum requirement for AGP 8.0+.
- **Key Capabilities**:
  - **Sealed Classes**: Controlled class hierarchies and exhaustive pattern matching.
  - **Pattern Matching for `instanceof`**: Eliminates redundant casting.
  - **Strong Encapsulation of JDK Internals**: Enforces clean API boundaries.
  - **Text Blocks**: Multiline string literals.

---

## 4. Combinations vs. Solos

### Dual-JDK Combinations (`25+21`, `25+17`, `21+17`)
- **Mechanism**: `JAVA_HOME` points to a fast host JDK (e.g. OpenJDK 25). In `~/.gradle/gradle.properties`, both host and toolchain paths are registered:
  ```properties
  org.gradle.java.installations.paths=/data/data/com.termux/files/usr/lib/jvm/java-25-openjdk,/data/data/com.termux/files/usr/lib/jvm/java-21-openjdk
  ```
- **Advantages**:
  - Maximum speed and responsiveness for shell tools and host utilities.
  - Zero bytecode warnings or AGP daemon failures during APK builds.
  - Automatic toolchain discovery by Gradle.
- **Trade-offs**: Requires keeping two OpenJDK packages installed (~500 MB to ~550 MB total storage).

### Single-JDK Solos (`21`, `25`, `17`)
- **Mechanism**: A single OpenJDK installation serves as both the Termux host environment and the Gradle compiler.
- **Advantages**:
  - **Storage Savings**: Consumes only ~220 MB to ~280 MB, saving ~300 MB of internal storage compared to combinations.
  - **Simplicity**: Single runtime in PATH, zero toolchain configuration overhead.
  - **Low RAM Impact**: Fewer packages and daemon processes on budget devices.
- **Trade-offs**: Terminal commands and build tools share the exact same runtime version.

---

## 5. Official Stable Profile Comparison Matrix

| Profile ID | Architecture | Host JDK | Gradle Toolchain | Disk Usage | Min. RAM | AGP 8.x/9.x Status | Repo Source | Primary Use Case |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `25+21` | Combination | OpenJDK 25 | OpenJDK 21 LTS | ~550 MB | 4 GB+ | 100% Compatible | Termux Official | **Dual-JDK Performance**: Fast 2025 LTS host with rock-solid Android builds |
| `25+17` | Combination | OpenJDK 25 | OpenJDK 17 LTS | ~500 MB | 4 GB+ | 100% (AGP 8.0) | Termux Official | Modern OpenJDK 25 host paired with Java 17 toolchain for older projects |
| `21+17` | Combination | OpenJDK 21 | OpenJDK 17 LTS | ~490 MB | 3 GB+ | 100% Compatible | Termux Official | LTS Java 21 host paired with legacy Java 17 compilation toolchain |
| `21` | Solo | OpenJDK 21 | OpenJDK 21 LTS | ~270 MB | 2 GB+ | 100% Compatible | Termux Official | **Recommended Default**: Saves ~300MB, lowest RAM overhead, zero bytecode conflicts, 100% stable with all AGP versions |
| `25` | Solo | OpenJDK 25 | OpenJDK 25 LTS | ~280 MB | 4 GB+ | Modern / Gradle 9+ | Termux Official | **Modern Solo**: Single OpenJDK 25 LTS runtime for cutting-edge projects |
| `17` | Solo | OpenJDK 17 | OpenJDK 17 LTS | ~220 MB | 2 GB+ | 100% (AGP 8.0) | Termux Official | **Legacy Baseline**: Saves ~330MB, maximum backward compatibility |

---

## 6. How to Switch Profiles

### Interactive TUI Chooser
Run:
```bash
tadev java switch
```
- Navigate with `[Up/Down]` or `j`/`k`.
- Select with `[Enter]`.
- Press `[Tab]` to open this comparison document directly in your device browser.
- Press `[Esc]` to cancel.

### Direct Command Line
Switch instantly to any profile ID:
```bash
# Switch to recommended default (OpenJDK 21 LTS Solo)
tadev java switch 21

# Switch to high-performance dual-JDK (OpenJDK 25 host + OpenJDK 21 toolchain)
tadev java switch 25+21

# Switch to modern OpenJDK 25 solo
tadev java switch 25

# Switch to legacy OpenJDK 17 solo
tadev java switch 17

# Switch to OpenJDK 25 host + OpenJDK 17 toolchain
tadev java switch 25+17
```

### Shell Shortcut
After environment setup, use the `use-java` shell function:
```bash
use-java 21
use-java 25+21
```
Running `use-java` without arguments prints the current configuration and lists all available profiles.
