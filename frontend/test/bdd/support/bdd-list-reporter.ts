import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import type {
  FullConfig,
  FullResult,
  Reporter,
  Suite,
  TestCase,
  TestResult,
} from '@playwright/test/reporter'

const supportDir = path.dirname(fileURLToPath(import.meta.url))
const bddConfigDir = path.resolve(supportDir, '..')
const projectRoot = path.resolve(bddConfigDir, '../../..')

const statusMarks: Record<string, string> = {
  passed: '✓',
  failed: '✘',
  timedOut: '✘',
  interrupted: '✘',
  skipped: '-',
}

export default class BddListReporter implements Reporter {
  private readonly printedFeatures = new Set<string>()
  private readonly featurePathByGeneratedFile = new Map<string, string>()
  private startTime = 0

  onBegin(_config: FullConfig, suite: Suite) {
    this.startTime = Date.now()
    const testCount = suite.allTests().length
    console.log(`Running ${testCount} browser BDD tests`)
    console.log('')
  }

  onTestEnd(test: TestCase, result: TestResult) {
    const featurePath = this.featurePath(test)
    if (!this.printedFeatures.has(featurePath)) {
      if (this.printedFeatures.size > 0) {
        console.log('')
      }
      this.printedFeatures.add(featurePath)
      console.log(featurePath)
    }

    const mark = statusMarks[result.status] ?? '?'
    const duration = this.formatDuration(result.duration)
    console.log(`  ${mark} ${test.title} (${duration})`)
    this.printErrors(result)
  }

  onEnd(result: FullResult) {
    const duration = this.formatDuration(Date.now() - this.startTime)
    console.log('')
    console.log(`${result.status.toUpperCase()} (${duration})`)
  }

  printsToStdio() {
    return true
  }

  private featurePath(test: TestCase) {
    const generatedFile = test.location.file
    const cached = this.featurePathByGeneratedFile.get(generatedFile)
    if (cached) {
      return cached
    }

    const featurePath = this.readGeneratedFeaturePath(generatedFile) ?? generatedFile
    this.featurePathByGeneratedFile.set(generatedFile, featurePath)
    return featurePath
  }

  private readGeneratedFeaturePath(generatedFile: string) {
    const content = fs.readFileSync(generatedFile, 'utf8')
    const match = content.match(/^\/\/ Generated from: (.+)$/m)
    if (!match) {
      return undefined
    }

    const absoluteFeaturePath = path.resolve(bddConfigDir, match[1])
    return path.relative(projectRoot, absoluteFeaturePath)
  }

  private formatDuration(durationMs: number) {
    if (durationMs < 1000) {
      return `${durationMs}ms`
    }
    return `${(durationMs / 1000).toFixed(1)}s`
  }

  private printErrors(result: TestResult) {
    const messages = result.errors
      .map((error) => error.message)
      .filter((message): message is string => Boolean(message))
    if (messages.length === 0 && result.error?.message) {
      messages.push(result.error.message)
    }

    for (const message of messages) {
      const lines = message.trim().split('\n')
      for (const line of lines) {
        console.log(`    ${line}`)
      }
    }
  }
}
