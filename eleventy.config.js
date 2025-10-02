export default async function(eleventyConfig) {
    eleventyConfig.setBrowserSyncConfig({
        files: './_site/**/*.css'
    });
    eleventyConfig.addPassthroughCopy('src/css');
};

export const config = {
    dir: {
        input: "src",
        output: "_site"
    }
};
