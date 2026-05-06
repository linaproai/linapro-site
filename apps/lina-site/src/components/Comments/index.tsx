import { useColorMode } from "@docusaurus/theme-common";
import Giscus from "@giscus/react";

// https://rikublock.dev/docs/tutorials/giscus-integration/
export default function Comments(): JSX.Element {
  const { colorMode } = useColorMode();

  return (
    <div className="docusaurus-mt-lg">
      <Giscus
        id="comments"
        data-repo="linaproai/linapro-site"
        data-repo-id="R_kgDOSIbhHw"
        category="General"
        categoryId="DIC_kwDOSIbhH84C8cwX"
        mapping="pathname"
        strict="1"
        reactionsEnabled="1"
        emitMetadata="0"
        inputPosition="top"
        theme={colorMode === "dark" ? "dark_tritanopia" : "light_tritanopia"}
        lang="en"
        loading="lazy"
      />
    </div>
  );
}
