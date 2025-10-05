import { eleventyImageTransformPlugin } from "@11ty/eleventy-img";

export default async function(eleventyConfig) {
    eleventyConfig.addPlugin(eleventyImageTransformPlugin);
    eleventyConfig.setBrowserSyncConfig({ files: './_site/**/*.css' });
    eleventyConfig.addPassthroughCopy({
        'src/css': 'css',
        'src/media': '/media',
        'src/assets': '/',
    });
};

export const config = {
    dir: {
        input: "src",
        output: "_site"
    }
};
