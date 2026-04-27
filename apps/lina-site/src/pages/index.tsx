import Link from '@docusaurus/Link';
import Translate, {translate} from '@docusaurus/Translate';
import useBaseUrl from '@docusaurus/useBaseUrl';
import CodeBlock from '@theme/CodeBlock';
import Layout from '@theme/Layout';

const REPO_URL = 'https://github.com/linaproai/linapro';
const DOCS_URL = '/quickstart';

function HomepageHeader() {
    const githubImage = useBaseUrl('/img/github.svg');
    const logoImage = useBaseUrl('/img/linapro-logo.png');
    return (
        <header className="home-section home-section--hero">
            <div className="container logo-container">
                <div>
                    <img
                        src={logoImage}
                        className="logo"
                        alt={translate({
                            id: 'home.hero.logoAlt',
                            message: 'LinaPro',
                            description: 'Alt text for the LinaPro logo in the hero',
                        })}
                    />
                </div>

                <h1 className="hero-title">
                    <Translate id="home.hero.title" description="Hero h1">
                        AI-native full-stack framework for sustainable delivery
                    </Translate>
                </h1>

                <p className="hero-lead">
                    <Translate id="home.hero.lead" description="Hero lead paragraph">
                        LinaPro makes AI a core engine of delivery: AI leads analysis, design, and implementation while teams set direction and make the critical calls.
                        With a core host service, admin workspace, plugin runtime, and spec-driven AI-native R&D workflow built in, teams can ship production-grade applications quickly while keeping architecture, testing, and governance ready to evolve.
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
                A general-purpose Go backend runtime that handles API contracts, service governance, auth, permissions, and the plugin lifecycle out of the box.
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
                A production-ready Vue 3 workspace that ships as the reference UI for every built-in capability — extend it, or swap it out entirely.
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
                Hot-loadable source plugins and WASM runtime plugins, each sandboxed with namespaced database and filesystem access — extend or replace anything without touching the host.
            </Translate>
        ),
    },
    {
        id: 'openspec',
        moduleKey: 'openspec/',
        title: (
            <Translate id="home.layer.openspec.title" description="Layer 4 title">
                AI-native R&D Workflow
            </Translate>
        ),
        desc: (
            <Translate id="home.layer.openspec.desc" description="Layer 4 description">
                A spec-first workflow that keeps AI, humans, and the codebase in lockstep across every iteration.
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
                        Loosely Coupled Architecture
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.layers.lead" description="Layers section lead">
                        Every layer is designed independently around a strict loose-coupling principle, so business modules can be turned on or off on demand without dragging the rest of the stack along for the ride.
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
                Map out the problem space and weigh the options before committing to one.
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
                Turn intent into an incremental spec, anchored to clear acceptance criteria.
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
                AI builds against the spec and writes the matching E2E tests alongside it.
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
                Humans sign off on the spec, the code, and the tests at the gates that matter.
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
                The verified spec joins the baseline, so the next iteration starts on solid ground.
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
                        Spec-driven AI development loop
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.workflow.lead" description="Workflow section lead">
                        Every iteration runs through a closed pipeline anchored to incremental spec files and mandatory E2E tests. AI always builds on a verified foundation — no architectural drift, no gaps in coverage.
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
                AI-native R&D workflow
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.workflow.desc" description="Strength 1 desc">
                Explore → Propose → Implement → Review → Archive. AI always moves forward from a verified spec baseline, so drift never gets a foothold.
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
                Every business module stands on its own and talks to others through interfaces, never hard dependencies — composable, swappable, and enabled on demand.
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
                Compile-time source plugins and runtime WASM plugins both run in isolated sandboxes, each with its own namespaced database and filesystem access.
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
                JWT auth plus declarative RBAC, with permissions expressed as struct tags — auditable by design. Millisecond permission propagation, sensitive-field masking, force-logout, and IP and device fingerprinting all included.
            </Translate>
        ),
    },
    {
        id: 'distributed',
        Icon: IconNetwork,
        title: (
            <Translate id="home.strength.distributed.title" description="Strength 5 title">
                Distributed by design
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.distributed.desc" description="Strength 5 desc">
                Permission topology sync, distributed locks, and the KV cache are all cluster-aware. Scale from a single node to a full cluster without rearchitecting a thing.
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
                Core services, official plugins, and a feature-rich management workspace, all in the box — focus on the business from day one instead of building infrastructure.
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
                        Framework Highlights
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.strengths.lead" description="Strengths section lead">
                        A framework engineered so AI productivity compounds with every iteration — instead of buckling under its own weight.
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

const previewSlots = [
    {
        id: 'overview',
        caption: (
            <Translate id="home.modules.preview.overview" description="Screenshot slot: workspace overview">
                Workspace overview
            </Translate>
        ),
    },
    {
        id: 'rbac',
        caption: (
            <Translate id="home.modules.preview.rbac" description="Screenshot slot: users & roles">
                Users & role permissions
            </Translate>
        ),
    },
    {
        id: 'modules',
        caption: (
            <Translate id="home.modules.preview.modules" description="Screenshot slot: business modules">
                Business modules
            </Translate>
        ),
    },
    {
        id: 'monitor',
        caption: (
            <Translate id="home.modules.preview.monitor" description="Screenshot slot: monitoring & audit">
                Monitoring & audit
            </Translate>
        ),
    },
];

function Modules() {
    return (
        <section className="home-section home-section--modules">
            <div className="container">
                <h2 className="section-title">
                    <Translate id="home.modules.title" description="Modules section title">
                        A workspace ready to use on day one
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.modules.lead" description="Modules section lead">
                        Build straight on top, extend any module through the plugin system, or replace one outright — all without touching the core host. Every module is wired into RBAC, and permission changes take effect on the spot, no re-login required.
                    </Translate>
                </p>
                <div className="modules-preview-grid">
                    {previewSlots.map((slot) => (
                        <div key={slot.id} className="modules-preview-card">
                            <div className="modules-preview-placeholder">
                                {slot.caption}
                            </div>
                        </div>
                    ))}
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
        label: 'INITIALIZE',
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
                        One-line installers for macOS, Linux, and Windows; a single Make entrypoint for every common task; and a default workspace ready to log into.
                    </Translate>
                </p>
                <div className="quickstart-blocks">
                    {quickStartBlocks.map((b) => (
                        <div key={b.id} className="quickstart-block">
                            <div className="quickstart-block-label">{b.label}</div>
                            <CodeBlock language="bash" className="quickstart-codeblock">
                                {b.body}
                            </CodeBlock>
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
                        Ready to ship AI-driven software?
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.cta.lead" description="Final CTA lead">
                        Read the docs, dig into the source, or jump straight into the workspace — LinaPro is open and waiting.
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
            <Strengths />
            <Layers />
            <Workflow />
            <Modules />
            <QuickStart />
            <FinalCta />
        </Layout>
    );
}
