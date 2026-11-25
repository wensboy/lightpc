import { GreetService as GreetBinding } from '../../bindings/github.com/wensboy/lightpc/backend/service/greet'

/** GreetService is the sample service the shell uses to exercise the log routes. */
export interface GreetService {
  /** greet answers with a greeting and writes one record through every route. */
  greet(name: string): Promise<string>
  /** emitLogs writes one record per level through the database entry, and answers how many it wrote. */
  emitLogs(): Promise<number>
}

export const greetService: GreetService = {
  async greet(name) {
    return String(await GreetBinding.Greet(name))
  },

  async emitLogs() {
    return Number(await GreetBinding.EmitLogs())
  },
}
