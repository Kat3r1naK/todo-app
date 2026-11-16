export class ThreadLock {
  private tails = new Map<string, Promise<unknown>>()

  async runExclusive<T>(key: string, fn: () => Promise<T>): Promise<T> {
    const previous = this.tails.get(key) ?? Promise.resolve()
    const current = previous.then(fn, fn)
    const tail = current.then(
      () => undefined,
      () => undefined,
    )
    this.tails.set(key, tail)
    try {
      return await current
    } finally {
      if (this.tails.get(key) === tail) {
        this.tails.delete(key)
      }
    }
  }
}
