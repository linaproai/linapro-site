import Link from '@docusaurus/Link';
import Translate, {translate} from '@docusaurus/Translate';
import useBaseUrl from '@docusaurus/useBaseUrl';
import Layout from '@theme/Layout';

const REPO_URL = 'https://github.com/linaproai/linapro';
const DOCS_URL = '/preface';

function HomepageHeader() {
    const githubImage = useBaseUrl('/img/github.svg');
    return (
        <header className="home-section home-section--hero">
            <div className="container logo-container">
                <div>
                    <img
                        src="/img/linapro-logo.png"
                        className="logo"
                        alt={translate({
                            id: 'home.hero.logoAlt',
                            message: 'LinaPro',
                            description: 'Alt text for the LinaPro logo in the hero',
                        })}
                    />
                    <div className="logo-badges">
                        <a href={REPO_URL} target="_blank" rel="noreferrer">
                            <img
                                src="https://img.shields.io/badge/production-ready-blue.svg"
                                alt="Production Ready"
                            />
                        </a>
                        <a href={REPO_URL} target="_blank" rel="noreferrer">
                            <img
                                src="https://img.shields.io/github/license/linaproai/linapro.svg?style=flat"
                                alt="License"
                            />
                        </a>
                        <a href={`${REPO_URL}/releases`} target="_blank" rel="noreferrer">
                            <img
                                src="https://img.shields.io/github/v/release/linaproai/linapro"
                                alt="Latest release"
                            />
                        </a>
                    </div>
                </div>

                <h1 className="hero-title">
                    <Translate id="home.hero.title" description="Hero h1">
                        AI-driven full-stack framework, engineered for sustainable delivery
                    </Translate>
                </h1>

                <p className="hero-lead">
                    <Translate id="home.hero.lead" description="Hero lead paragraph">
                        LinaPro is an AI-driven full-stack development framework where AI handles the bulk of analysis, design, and implementation, while humans set direction and own quality at the gates that matter. A core host service, a production-ready management workspace, a dual-mode plugin runtime, and a specification-driven AI R&D workflow lock together so delivery quality keeps pace as the product grows.
                    </Translate>
                </p>

                <div className="hero-cta">
                    <Link
                        className="button button--primary button--md"
                        to={DOCS_URL}
                        style={{width: '200px'}}
                    >
                        <Translate id="home.hero.cta.start" description="Hero primary CTA">
                            Get Started →
                        </Translate>
                    </Link>
                    <Link
                        className="button button--secondary button--md hover:bg-gray-200"
                        to={REPO_URL}
                        style={{
                            width: '200px',
                            paddingLeft: '50px',
                            backgroundImage: `url(${githubImage})`,
                            backgroundRepeat: 'no-repeat',
                            backgroundPosition: '45px center',
                        }}
                    >
                        <Translate id="home.hero.cta.github" description="Hero secondary CTA">
                            GitHub →
                        </Translate>
                    </Link>
                </div>
            </div>
        </header>
    );
}

const layers = [
    {
        id: 'core',
        moduleKey: 'apps/lina-core',
        title: (
            <Translate id="home.layer.core.title" description="Layer 1 title">
                Core Host Service
            </Translate>
        ),
        desc: (
            <Translate id="home.layer.core.desc" description="Layer 1 description">
                Universal Go backend runtime — API contracts, service governance, authentication, permissions, plugin lifecycle.
            </Translate>
        ),
    },
    {
        id: 'vben',
        moduleKey: 'apps/lina-vben',
        title: (
            <Translate id="home.layer.vben.title" description="Layer 2 title">
                Management Workspace
            </Translate>
        ),
        desc: (
            <Translate id="home.layer.vben.desc" description="Layer 2 description">
                Production-ready Vue 3 frontend workspace — the reference UI for every built-in capability, ready to extend or replace.
            </Translate>
        ),
    },
    {
        id: 'plugins',
        moduleKey: 'apps/lina-plugins',
        title: (
            <Translate id="home.layer.plugins.title" description="Layer 3 title">
                Dual-mode Plugin System
            </Translate>
        ),
        desc: (
            <Translate id="home.layer.plugins.desc" description="Layer 3 description">
                Hot-loadable source and WASM dynamic plugins, sandboxed with namespaced database and filesystem access — extend or replace anything without touching the host.
            </Translate>
        ),
    },
    {
        id: 'openspec',
        moduleKey: 'openspec/',
        title: (
            <Translate id="home.layer.openspec.title" description="Layer 4 title">
                AI R&D Workflow
            </Translate>
        ),
        desc: (
            <Translate id="home.layer.openspec.desc" description="Layer 4 description">
                Specification-first workflow that keeps AI, humans, and the codebase aligned across every iteration.
            </Translate>
        ),
    },
];

function Layers() {
    return (
        <section className="home-section home-section--layers">
            <div className="container">
                <h2 className="section-title">
                    <Translate id="home.layers.title" description="Layers section title">
                        Built on four interlocking layers
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.layers.lead" description="Layers section lead">
                        Each layer is designed independently under a strict loose-coupling principle, so business modules can be enabled or disabled on demand without dragging the rest of the stack along.
                    </Translate>
                </p>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-5 mt-8">
                    {layers.map((layer, idx) => (
                        <div key={layer.id} className="card layer-card p-8 box-border">
                            <div className="layer-number">{`0${idx + 1}`}</div>
                            <h3 className="layer-title">{layer.title}</h3>
                            <code className="layer-key">{layer.moduleKey}</code>
                            <p className="layer-desc">{layer.desc}</p>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
}

const workflowSteps = [
    {
        id: 'explore',
        title: (
            <Translate id="home.workflow.explore.title" description="Workflow stage 1">
                Explore
            </Translate>
        ),
        desc: (
            <Translate id="home.workflow.explore.desc" description="Workflow stage 1 desc">
                Investigate the problem space and surface options before committing.
            </Translate>
        ),
    },
    {
        id: 'propose',
        title: (
            <Translate id="home.workflow.propose.title" description="Workflow stage 2">
                Propose
            </Translate>
        ),
        desc: (
            <Translate id="home.workflow.propose.desc" description="Workflow stage 2 desc">
                Capture intent as an incremental specification anchored to acceptance criteria.
            </Translate>
        ),
    },
    {
        id: 'implement',
        title: (
            <Translate id="home.workflow.implement.title" description="Workflow stage 3">
                Implement
            </Translate>
        ),
        desc: (
            <Translate id="home.workflow.implement.desc" description="Workflow stage 3 desc">
                AI executes against the spec and writes the corresponding E2E tests.
            </Translate>
        ),
    },
    {
        id: 'review',
        title: (
            <Translate id="home.workflow.review.title" description="Workflow stage 4">
                Review
            </Translate>
        ),
        desc: (
            <Translate id="home.workflow.review.desc" description="Workflow stage 4 desc">
                Humans review the spec, the code, and the tests at the gate that matters.
            </Translate>
        ),
    },
    {
        id: 'archive',
        title: (
            <Translate id="home.workflow.archive.title" description="Workflow stage 5">
                Archive
            </Translate>
        ),
        desc: (
            <Translate id="home.workflow.archive.desc" description="Workflow stage 5 desc">
                The verified spec joins the baseline — the next iteration builds on solid ground.
            </Translate>
        ),
    },
];

function Workflow() {
    return (
        <section className="home-section home-section--workflow">
            <div className="container">
                <h2 className="section-title">
                    <Translate id="home.workflow.title" description="Workflow section title">
                        A specification-driven AI R&D loop
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.workflow.lead" description="Workflow section lead">
                        Every iteration follows a closed pipeline anchored to incremental specification files and mandatory E2E tests. AI always builds on a verified foundation — no architectural drift, no test voids.
                    </Translate>
                </p>
                <div className="workflow-steps">
                    {workflowSteps.map((step, i) => (
                        <div key={step.id} className="workflow-step">
                            <div className="workflow-step-number">{`STAGE 0${i + 1}`}</div>
                            <h4>{step.title}</h4>
                            <p>{step.desc}</p>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
}

const IconSparkles = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 3l1.6 4.4L18 9l-4.4 1.6L12 15l-1.6-4.4L6 9l4.4-1.6z" />
        <path d="M19 15l.7 1.8L21.5 17.5l-1.8.7L19 20l-.7-1.8-1.8-.7 1.8-.7z" />
        <path d="M5 15l.5 1.3L6.8 16.8l-1.3.5L5 18.5l-.5-1.3L3.3 16.8l1.3-.5z" />
    </svg>
);

const IconBlocks = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="3" width="7" height="7" rx="1" />
        <rect x="14" y="3" width="7" height="7" rx="1" />
        <rect x="3" y="14" width="7" height="7" rx="1" />
        <rect x="14" y="14" width="7" height="7" rx="1" />
    </svg>
);

const IconPlug = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 2v6" />
        <path d="M15 2v6" />
        <path d="M5 8h14v3a5 5 0 0 1-5 5h-4a5 5 0 0 1-5-5z" />
        <path d="M12 16v6" />
    </svg>
);

const IconShield = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 3l8 3v6c0 5-3.5 8.5-8 9-4.5-.5-8-4-8-9V6z" />
        <path d="M9 12l2 2 4-4" />
    </svg>
);

const IconNetwork = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="5" r="2.2" />
        <circle cx="5" cy="19" r="2.2" />
        <circle cx="19" cy="19" r="2.2" />
        <path d="M12 7v3M12 10l-5.5 7M12 10l5.5 7" />
    </svg>
);

const IconCheckCircle = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="9" />
        <path d="M8 12l3 3 5-6" />
    </svg>
);

const strengths = [
    {
        id: 'workflow',
        Icon: IconSparkles,
        title: (
            <Translate id="home.strength.workflow.title" description="Strength 1 title">
                AI-native closed-loop R&D
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.workflow.desc" description="Strength 1 desc">
                Explore → Propose → Implement → Review → Archive. AI always advances from a verified specification baseline, eliminating drift before it starts.
            </Translate>
        ),
    },
    {
        id: 'coupling',
        Icon: IconBlocks,
        title: (
            <Translate id="home.strength.coupling.title" description="Strength 2 title">
                Module-level loose coupling
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.coupling.desc" description="Strength 2 desc">
                Every business module is independent and collaborates through interfaces, never hard dependencies — composable, replaceable, on-demand.
            </Translate>
        ),
    },
    {
        id: 'plugins',
        Icon: IconPlug,
        title: (
            <Translate id="home.strength.plugins.title" description="Strength 3 title">
                Dual-mode pluggable runtime
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.plugins.desc" description="Strength 3 desc">
                Compile-time source plugins and runtime WASM dynamic plugins run in isolated sandboxes with namespaced database and filesystem access.
            </Translate>
        ),
    },
    {
        id: 'governance',
        Icon: IconShield,
        title: (
            <Translate id="home.strength.governance.title" description="Strength 4 title">
                Enterprise-grade governance
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.governance.desc" description="Strength 4 desc">
                JWT auth + declarative RBAC permissions as struct tags — auditable by design. Millisecond-propagation, sensitive-field masking, force-logout, IP & device fingerprints.
            </Translate>
        ),
    },
    {
        id: 'distributed',
        Icon: IconNetwork,
        title: (
            <Translate id="home.strength.distributed.title" description="Strength 5 title">
                Distribution-ready by design
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.distributed.desc" description="Strength 5 desc">
                Permission topology sync, distributed locking, and key-value cache are cluster-aware. Scale from single node to multi-node with no architectural changes.
            </Translate>
        ),
    },
    {
        id: 'builtin',
        Icon: IconCheckCircle,
        title: (
            <Translate id="home.strength.builtin.title" description="Strength 6 title">
                Production-ready out of the box
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.builtin.desc" description="Strength 6 desc">
                Built-in core services, official plugins, and a rich management workspace — focus on business from day one instead of building infrastructure.
            </Translate>
        ),
    },
];

function Strengths() {
    return (
        <section className="home-section home-section--features">
            <div className="container">
                <h2 className="section-title">
                    <Translate id="home.strengths.title" description="Strengths section title">
                        Why teams choose LinaPro
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.strengths.lead" description="Strengths section lead">
                        A framework engineered so AI productivity compounds with every iteration instead of decaying under its own weight.
                    </Translate>
                </p>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-5 mt-8">
                    {strengths.map(({id, Icon, title, desc}) => (
                        <div key={id} className="card p-8 box-border">
                            <div className="feature-icon">
                                <Icon />
                            </div>
                            <h3 className="strength-title">{title}</h3>
                            <p className="strength-desc">{desc}</p>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
}

const builtinModules = [
    {key: 'user', label: <Translate id="home.module.user" description="Built-in module: User Management">User Management</Translate>},
    {key: 'role', label: <Translate id="home.module.role" description="Built-in module: Role Management">Role Management</Translate>},
    {key: 'menu', label: <Translate id="home.module.menu" description="Built-in module: Menu Management">Menu Management</Translate>},
    {key: 'dictionary', label: <Translate id="home.module.dictionary" description="Built-in module: Data Dictionary">Data Dictionary</Translate>},
    {key: 'parameters', label: <Translate id="home.module.parameters" description="Built-in module: Parameter Settings">Parameter Settings</Translate>},
    {key: 'files', label: <Translate id="home.module.files" description="Built-in module: File Management">File Management</Translate>},
    {key: 'jobs', label: <Translate id="home.module.jobs" description="Built-in module: Job Scheduling">Job Scheduling</Translate>},
    {key: 'plugins', label: <Translate id="home.module.plugins" description="Built-in module: Plugin Management">Plugin Management</Translate>},
    {key: 'apidocs', label: <Translate id="home.module.apidocs" description="Built-in module: API Docs">API Docs</Translate>},
    {key: 'sysinfo', label: <Translate id="home.module.sysinfo" description="Built-in module: System Info">System Info</Translate>},
];

const pluginModules = [
    {key: 'org-center', label: <Translate id="home.plugin.org" description="Plugin: org-center">Department & Position</Translate>},
    {key: 'content-notice', label: <Translate id="home.plugin.notice" description="Plugin: content-notice">Notice Management</Translate>},
    {key: 'monitor-online', label: <Translate id="home.plugin.online" description="Plugin: monitor-online">Online Sessions</Translate>},
    {key: 'monitor-server', label: <Translate id="home.plugin.server" description="Plugin: monitor-server">Server Monitor</Translate>},
    {key: 'monitor-operlog', label: <Translate id="home.plugin.operlog" description="Plugin: monitor-operlog">Operation Audit</Translate>},
    {key: 'monitor-loginlog', label: <Translate id="home.plugin.loginlog" description="Plugin: monitor-loginlog">Login Audit</Translate>},
];

function Modules() {
    return (
        <section className="home-section home-section--modules">
            <div className="container">
                <h2 className="section-title">
                    <Translate id="home.modules.title" description="Modules section title">
                        A workspace that ships ready to use
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.modules.lead" description="Modules section lead">
                        Build directly on top, extend any module via the plugin system, or replace one entirely — all without touching the core host. Every module is wired into RBAC; permission changes take effect without re-login.
                    </Translate>
                </p>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mt-8">
                    <div className="modules-column">
                        <h3>
                            <Translate id="home.modules.builtin.heading" description="Built-in modules heading">
                                Ships with
                            </Translate>{' '}
                            <code>lina-vben</code>
                        </h3>
                        <ul className="modules-list">
                            {builtinModules.map((m) => (
                                <li key={m.key} className="modules-list-item">
                                    <code>{m.key}</code>
                                    <span>{m.label}</span>
                                </li>
                            ))}
                        </ul>
                    </div>
                    <div className="modules-column">
                        <h3>
                            <Translate id="home.modules.plugins.heading" description="Plugin modules heading">
                                Via official plugins
                            </Translate>
                        </h3>
                        <ul className="modules-list">
                            {pluginModules.map((m) => (
                                <li key={m.key} className="modules-list-item">
                                    <code>{m.key}</code>
                                    <span>{m.label}</span>
                                </li>
                            ))}
                        </ul>
                    </div>
                </div>
            </div>
        </section>
    );
}

const quickStartBlocks = [
    {
        id: 'install',
        label: 'INSTALL',
        body: 'curl -fsSL https://linapro.ai/install.sh | bash',
    },
    {
        id: 'init',
        label: 'INITIALISE',
        body: 'make init confirm=init\nmake mock confirm=mock   # optional demo data',
    },
    {
        id: 'dev',
        label: 'DEVELOP',
        body: 'make dev\n# Workspace: http://localhost:5666\n# API:       http://localhost:8080',
    },
];

function QuickStart() {
    return (
        <section className="home-section home-section--quickstart">
            <div className="container">
                <h2 className="section-title">
                    <Translate id="home.quickstart.title" description="QuickStart section title">
                        Up and running in three commands
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.quickstart.lead" description="QuickStart section lead">
                        macOS, Linux, and Windows installers; a single Make entrypoint for every common task; a default workspace ready to log into.
                    </Translate>
                </p>
                <div className="quickstart-blocks">
                    {quickStartBlocks.map((b) => (
                        <div key={b.id} className="quickstart-block">
                            <div className="quickstart-block-label">{b.label}</div>
                            <pre>{b.body}</pre>
                        </div>
                    ))}
                </div>
                <div className="quickstart-footnote">
                    <Translate id="home.quickstart.footnote" description="QuickStart footnote">
                        Default account:
                    </Translate>{' '}
                    <code>admin</code> / <code>admin123</code>
                </div>
            </div>
        </section>
    );
}

function FinalCta() {
    const githubImage = useBaseUrl('/img/github.svg');
    return (
        <section className="home-section home-section--cta">
            <div className="container text-center">
                <h2 className="section-title">
                    <Translate id="home.cta.title" description="Final CTA title">
                        Ready to ship AI-native software?
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.cta.lead" description="Final CTA lead">
                        Read the docs, dive into the source, or jump straight into the workspace — LinaPro is open and ready.
                    </Translate>
                </p>
                <div className="cta-buttons">
                    <Link className="button button--primary button--md" to={DOCS_URL} style={{width: '200px'}}>
                        <Translate id="home.cta.docs" description="Final CTA docs button">
                            Read the Docs →
                        </Translate>
                    </Link>
                    <Link
                        className="button button--secondary button--md hover:bg-gray-200"
                        to={REPO_URL}
                        style={{
                            width: '200px',
                            paddingLeft: '50px',
                            backgroundImage: `url(${githubImage})`,
                            backgroundRepeat: 'no-repeat',
                            backgroundPosition: '22px center',
                        }}
                    >
                        <Translate id="home.cta.github" description="Final CTA github button">
                            Star on GitHub
                        </Translate>
                    </Link>
                </div>
            </div>
        </section>
    );
}

export default function Home(): JSX.Element {
    return (
        <Layout>
            <HomepageHeader />
            <Layers />
            <Workflow />
            <Strengths />
            <Modules />
            <QuickStart />
            <FinalCta />
        </Layout>
    );
}
