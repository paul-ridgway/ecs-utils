import ora, { Ora } from 'ora';

export function spinnerWrapper<T>(text: string, fn: (spinner: Ora) => Promise<T>): Promise<T> {
  const spinner = ora({
    color: 'yellow'
  }).start(text);
  return fn(spinner).then((r: T): T => {
    spinner.stopAndPersist();
    return r;
  });
}