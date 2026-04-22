import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

import styles from './pageShell.module.css';

const content = {
  en: {
    eyebrow: 'Reserved entry',
    title: 'Login will connect to the LinaPro user system in a later phase.',
    body: 'This route is intentionally public now so the information architecture stays stable, but the real authentication flow is not part of phase one.',
    status: 'placeholder',
    panelTitle: 'What phase one includes',
    lines: [
      'A stable public route',
      'Clear product expectation for future identity integration',
      'No fake form flow and no pseudo-auth behavior',
    ],
    primary: 'Read the docs',
    secondary: 'Open GitHub',
  },
  'zh-CN': {
    eyebrow: '预留入口',
    title: '登录能力将在后续阶段对接 LinaPro 用户系统。',
    body: '这个路由现在公开保留，是为了让站点信息架构保持稳定；真实认证流程不属于一期实现范围。',
    status: '占位中',
    panelTitle: '一期包含的内容',
    lines: [
      '稳定的公开路由',
      '对未来身份集成的清晰产品预期',
      '不伪造表单流程，也不模拟登录行为',
    ],
    primary: '阅读文档',
    secondary: '打开 GitHub',
  },
} as const;

export default function LoginPage(): ReactNode {
  const {i18n} = useDocusaurusContext();
  const locale = i18n.currentLocale as keyof typeof content;
  const copy = content[locale] ?? content.en;

  return (
    <Layout title={copy.eyebrow} description={copy.body}>
      <main className={styles.shell}>
        <div className={styles.inner}>
          <section className={styles.hero}>
            <div>
              <p className={styles.eyebrow}>{copy.eyebrow}</p>
              <h1>{copy.title}</h1>
              <p>{copy.body}</p>
              <div className={styles.actions}>
                <Link className={clsx('button', styles.primaryButton)} to="/docs">
                  {copy.primary}
                </Link>
                <Link
                  className={clsx('button', styles.secondaryButton)}
                  href="https://github.com/linaproai/linapro">
                  {copy.secondary}
                </Link>
              </div>
            </div>
            <aside className={styles.panel}>
              <span className={styles.status}>{copy.status}</span>
              <h2>{copy.panelTitle}</h2>
              {copy.lines.map((line) => (
                <div className={styles.infoLine} key={line}>
                  {line}
                </div>
              ))}
            </aside>
          </section>
        </div>
      </main>
    </Layout>
  );
}
