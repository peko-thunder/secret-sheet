import { spawn } from "child_process";

/**
 * Spawn a subprocess with the given env vars merged on top of the current
 * process environment, then forward its exit code.
 *
 * stdio is inherited so the child's output goes directly to the terminal.
 */
export function runCommand(
  command: string,
  args: string[],
  envVars: Record<string, string>
): Promise<void> {
  return new Promise((resolve, reject) => {
    const env: NodeJS.ProcessEnv = {
      ...process.env,
      ...envVars,
    };

    const child = spawn(command, args, {
      env,
      stdio: "inherit",
      // Use shell: false so secrets never appear in shell history
      shell: false,
    });

    child.on("error", (err) => {
      if ((err as NodeJS.ErrnoException).code === "ENOENT") {
        reject(new Error(`Command not found: ${command}`));
      } else {
        reject(err);
      }
    });

    child.on("close", (code, signal) => {
      if (signal) {
        // Mirror the signal to the parent process
        process.kill(process.pid, signal);
      } else {
        process.exitCode = code ?? 0;
        resolve();
      }
    });
  });
}
