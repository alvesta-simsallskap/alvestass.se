import { glob } from 'astro/loaders';
import { defineCollection, z } from 'astro:content';

const swimSchool = defineCollection({
  loader: glob({ base: './src/content/swim-school', pattern: '**/*.md' }),
  schema: z.object({
    title: z.string(),
    order: z.number(),
    image: z.string().optional(),
    imageAttribution: z.string().optional(),
  }),
});

const trainingGroups = defineCollection({
  loader: glob({ base: './src/content/training-groups', pattern: '**/*.md' }),
  schema: z.object({
    title: z.string(),
    image: z.string().optional(),
    imageAttribution: z.string().optional(),
  }),
});

const clubInfo = defineCollection({
  loader: glob({ base: './src/content/club', pattern: '**/*.md' }),
  schema: z.object({
    title: z.string(),
    image: z.string().optional(),
    imageAttribution: z.string().optional(),
  }),
});

export const collections = { swimSchool, trainingGroups, clubInfo };