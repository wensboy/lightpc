import { GreetService } from '../../bindings/github.com/wensboy/lightpc/backend/service/greet'

// greet wraps the Wails binding so feature code never touches generated bindings directly.
export function greet(name: string): Promise<string> {
  return GreetService.Greet(name)
}
