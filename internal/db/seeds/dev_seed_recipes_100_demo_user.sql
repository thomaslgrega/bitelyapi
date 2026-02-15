-- DEV / DEMO SEED DATA (100 recipes + ingredients)
-- Re-runnable: upserts recipes and replaces ingredients for these recipe IDs.
-- Uses a demo user selected dynamically by email.

\set ON_ERROR_STOP on

BEGIN;

-- Fail fast if the demo user does not exist
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'example@gmail.com') THEN
    RAISE EXCEPTION 'Seed user not found: users.email = example@gmail.com';
  END IF;
END $$;

-- Keep IDs in a temp table so we can reuse them across statements
CREATE TEMP TABLE seed_recipes (
  id uuid PRIMARY KEY,
  name text NOT NULL,
  category text NOT NULL,
  instructions text,
  thumbnail_url text,
  calories integer,
  total_cook_time integer
) ON COMMIT DROP;

INSERT INTO seed_recipes (id, name, category, instructions, thumbnail_url, calories, total_cook_time)
VALUES
('ced3ffc1-d43b-576c-8a7c-5579e64647aa'::uuid, 'Weeknight Beef Tacos', 'Beef', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/beef-weeknight-beef-tacos/800/500', 827, 27),
('5c584359-337b-5e7d-80c6-dd385f6881ec'::uuid, 'Beef & Spicy Stir-Fry', 'Beef', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/beef-beef--spicy-stir-fry/800/500', 512, 67),
('2cfeec0f-31a7-5b27-bf77-8e5e6563da84'::uuid, 'Classic Beef Chili', 'Beef', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/beef-classic-beef-chili/800/500', 640, 35),
('e410dd2f-f321-564c-9cec-a3cfb036c9c0'::uuid, 'Beef Burger (Lemon)', 'Beef', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/beef-beef-burger-lemon/800/500', 614, 28),
('133b2dc5-06f4-5ff3-9eef-c3a9370c0680'::uuid, 'Beef Rice Bowl (Herb)', 'Beef', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/beef-beef-rice-bowl-herb/800/500', 877, 26),
('ef51d7ce-9ecd-50fe-b653-c951faf008e8'::uuid, 'Beef Lettuce Wraps', 'Beef', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/beef-beef-lettuce-wraps/800/500', 846, 67),
('cc80dc63-8e54-5cf7-821c-97961e58f84e'::uuid, 'Beef Sweet Skillet', 'Beef', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/beef-beef-sweet-skillet/800/500', 779, 25),
('9cdbc9bb-02c3-552d-9e96-36036f40e948'::uuid, 'Beef Pasta Bake', 'Beef', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/beef-beef-pasta-bake/800/500', 802, 47),
('2866b2a1-3eb7-506b-9065-ae7c861d5de1'::uuid, 'Slow-Simmer Beef Stew', 'Beef', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/beef-slow-simmer-beef-stew/800/500', 516, 21),
('558ac48b-627e-5b8e-afb6-a88d05cbda69'::uuid, 'Beef Quesadillas', 'Beef', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/beef-beef-quesadillas/800/500', 547, 33),
('6e70f710-b1e7-5747-86d7-85cd61e9d648'::uuid, 'Lemon Garlic Chicken', 'Chicken', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/chicken-lemon-garlic-chicken/800/500', 569, 47),
('e22aaba5-57ed-5a58-83f9-bd3ec6c4c178'::uuid, 'Chicken Fajita Bowls', 'Chicken', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/chicken-chicken-fajita-bowls/800/500', 758, 16),
('9d7b95e2-2923-5160-aff6-03de0f3dc88b'::uuid, 'Chicken Teriyaki', 'Chicken', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/chicken-chicken-teriyaki/800/500', 737, 27),
('634d8556-441d-5180-90a4-347a256e4540'::uuid, 'Creamy Chicken Pasta', 'Chicken', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/chicken-creamy-chicken-pasta/800/500', 816, 56),
('49272eb3-1b6c-5cd3-859d-5c0a6e2bda4f'::uuid, 'Chicken Smoky Salad', 'Chicken', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/chicken-chicken-smoky-salad/800/500', 809, 49),
('a63fbab2-fa52-5a49-a507-ef7f9f9e62cf'::uuid, 'Sheet Pan Chicken & Veggies', 'Chicken', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/chicken-sheet-pan-chicken--veggies/800/500', 664, 29),
('19aeb489-becf-5485-adc7-f6d12917032c'::uuid, 'Chicken Stir-Fry (Crispy)', 'Chicken', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/chicken-chicken-stir-fry-crispy/800/500', 679, 52),
('e1b6903b-3107-58e3-a305-b17355f68c25'::uuid, 'Chicken Parmesan (Quick)', 'Chicken', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/chicken-chicken-parmesan-quick/800/500', 592, 15),
('d80397f2-d891-5bff-b189-a4f8746ac62d'::uuid, 'Chicken Tortilla Soup', 'Chicken', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/chicken-chicken-tortilla-soup/800/500', 838, 25),
('ada2dc0d-afd7-5613-8d26-fba3697b5ceb'::uuid, 'Honey Mustard Chicken', 'Chicken', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/chicken-honey-mustard-chicken/800/500', 807, 42),
('0bc39706-d1d5-57fa-b5bc-d2fcd21f75ef'::uuid, 'Fudgy Cocoa Brownies', 'Dessert', $$1) Preheat oven (if baking) and prepare your pan.
2) Mix wet ingredients, then fold in dry ingredients just until combined.
3) Bake/chill until set; cool before slicing or serving.
4) Store leftovers covered.$$, 'https://picsum.photos/seed/dessert-fudgy-cocoa-brownies/800/500', 267, 37),
('7c2e30ef-cede-58f6-b238-10adbbf4c0be'::uuid, 'Chocolate Chip Cookies', 'Dessert', $$1) Preheat oven (if baking) and prepare your pan.
2) Mix wet ingredients, then fold in dry ingredients just until combined.
3) Bake/chill until set; cool before slicing or serving.
4) Store leftovers covered.$$, 'https://picsum.photos/seed/dessert-chocolate-chip-cookies/800/500', 219, 33),
('98746583-0d9f-5278-a14c-fc7b545ab8b8'::uuid, 'Classic Banana Bread', 'Dessert', $$1) Preheat oven (if baking) and prepare your pan.
2) Mix wet ingredients, then fold in dry ingredients just until combined.
3) Bake/chill until set; cool before slicing or serving.
4) Store leftovers covered.$$, 'https://picsum.photos/seed/dessert-classic-banana-bread/800/500', 375, 41),
('c0b3a766-aa45-5938-95cd-a747d75afc78'::uuid, 'Smoky Mug Cake', 'Dessert', $$1) Preheat oven (if baking) and prepare your pan.
2) Mix wet ingredients, then fold in dry ingredients just until combined.
3) Bake/chill until set; cool before slicing or serving.
4) Store leftovers covered.$$, 'https://picsum.photos/seed/dessert-smoky-mug-cake/800/500', 206, 25),
('55f27fee-fbb9-5a13-8c36-f122413f5fcc'::uuid, 'No-Bake Cheesecake Cups (Sweet)', 'Dessert', $$1) Preheat oven (if baking) and prepare your pan.
2) Mix wet ingredients, then fold in dry ingredients just until combined.
3) Bake/chill until set; cool before slicing or serving.
4) Store leftovers covered.$$, 'https://picsum.photos/seed/dessert-no-bake-cheesecake-cups-sweet/800/500', 277, 26),
('8d90b5b8-9a45-5623-9006-5b2e4b73dd68'::uuid, 'Oatmeal Raisin Cookies', 'Dessert', $$1) Preheat oven (if baking) and prepare your pan.
2) Mix wet ingredients, then fold in dry ingredients just until combined.
3) Bake/chill until set; cool before slicing or serving.
4) Store leftovers covered.$$, 'https://picsum.photos/seed/dessert-oatmeal-raisin-cookies/800/500', 271, 74),
('3bd1c196-f1c9-580f-bd62-42e17cf162da'::uuid, 'Lemon Bars (Creamy)', 'Dessert', $$1) Preheat oven (if baking) and prepare your pan.
2) Mix wet ingredients, then fold in dry ingredients just until combined.
3) Bake/chill until set; cool before slicing or serving.
4) Store leftovers covered.$$, 'https://picsum.photos/seed/dessert-lemon-bars-creamy/800/500', 268, 58),
('c4e30e25-3b9e-518c-a699-5d02cec395ae'::uuid, 'Apple Crisp (Zesty)', 'Dessert', $$1) Preheat oven (if baking) and prepare your pan.
2) Mix wet ingredients, then fold in dry ingredients just until combined.
3) Bake/chill until set; cool before slicing or serving.
4) Store leftovers covered.$$, 'https://picsum.photos/seed/dessert-apple-crisp-zesty/800/500', 247, 71),
('066df4b8-d7d8-5cb4-9438-4f72c9c903bb'::uuid, 'Peanut Butter Cookies', 'Dessert', $$1) Preheat oven (if baking) and prepare your pan.
2) Mix wet ingredients, then fold in dry ingredients just until combined.
3) Bake/chill until set; cool before slicing or serving.
4) Store leftovers covered.$$, 'https://picsum.photos/seed/dessert-peanut-butter-cookies/800/500', 191, 66),
('bcd395fe-4f1e-52ab-8c25-9e16dc22020a'::uuid, 'Chocolate Pudding (Ginger)', 'Dessert', $$1) Preheat oven (if baking) and prepare your pan.
2) Mix wet ingredients, then fold in dry ingredients just until combined.
3) Bake/chill until set; cool before slicing or serving.
4) Store leftovers covered.$$, 'https://picsum.photos/seed/dessert-chocolate-pudding-ginger/800/500', 297, 54),
('2981d71a-8174-569f-beaf-f8e601792726'::uuid, 'Creamy Tomato Soup', 'Other', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/other-creamy-tomato-soup/800/500', 277, 29),
('9a6490b0-e2e5-5aea-bc00-46d3a216e5c5'::uuid, 'Turkey Chili (Quick)', 'Other', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/other-turkey-chili-quick/800/500', 230, 40),
('95714ebb-ae76-5fb3-a3da-882595b267fa'::uuid, 'Veggie & Rice Bowl', 'Other', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/other-veggie--rice-bowl/800/500', 450, 58),
('501955a6-71c6-5fb7-83b3-d68f9b714477'::uuid, 'Grilled Cheese (Sweet)', 'Other', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/other-grilled-cheese-sweet/800/500', 520, 41),
('43b4da68-85b9-5379-831c-bd5d8dd4775b'::uuid, 'Simple Fried Rice', 'Other', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/other-simple-fried-rice/800/500', 346, 50),
('0ab19725-1a7c-5d79-9c96-799e59098665'::uuid, 'Quinoa Bowl (Creamy)', 'Other', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/other-quinoa-bowl-creamy/800/500', 221, 7),
('0ae2d4b1-423d-5415-bea9-2b9d08fb5a22'::uuid, 'Baked Potatoes (Zesty)', 'Other', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/other-baked-potatoes-zesty/800/500', 383, 54),
('39df1c43-555e-53e5-9414-b2bb3a88237b'::uuid, 'Homemade Salsa (Sesame)', 'Other', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/other-homemade-salsa-sesame/800/500', 446, 10),
('3b23f590-17da-51e6-8aff-c378e559841c'::uuid, 'Easy Hummus (Ginger)', 'Other', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/other-easy-hummus-ginger/800/500', 388, 60),
('a176f573-5785-5cd4-bd4a-c7ad0d0b2469'::uuid, 'Snack Plate (BBQ)', 'Other', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/other-snack-plate-bbq/800/500', 253, 29),
('fdd74c62-5009-53d9-bc37-5203f4a6fe6d'::uuid, 'Tomato Basil Pasta', 'Pasta', $$1) Boil pasta in salted water until al dente; reserve a splash of pasta water.
2) Build sauce in a pan, then toss pasta with sauce and a little pasta water.
3) Finish with cheese/herbs and adjust seasoning.$$, 'https://picsum.photos/seed/pasta-tomato-basil-pasta/800/500', 592, 44),
('7bc5b32b-6e41-562b-a1e0-57acb7e52248'::uuid, 'Pesto Pasta with Smoky', 'Pasta', $$1) Boil pasta in salted water until al dente; reserve a splash of pasta water.
2) Build sauce in a pan, then toss pasta with sauce and a little pasta water.
3) Finish with cheese/herbs and adjust seasoning.$$, 'https://picsum.photos/seed/pasta-pesto-pasta-with-smoky/800/500', 775, 38),
('23f9d62b-3156-58eb-8e9b-c570307482cb'::uuid, 'Baked Ziti', 'Pasta', $$1) Boil pasta in salted water until al dente; reserve a splash of pasta water.
2) Build sauce in a pan, then toss pasta with sauce and a little pasta water.
3) Finish with cheese/herbs and adjust seasoning.$$, 'https://picsum.photos/seed/pasta-baked-ziti/800/500', 533, 38),
('6f4ee5d9-05df-5a2b-955b-2e75877e2539'::uuid, 'Garlic Parmesan Pasta', 'Pasta', $$1) Boil pasta in salted water until al dente; reserve a splash of pasta water.
2) Build sauce in a pan, then toss pasta with sauce and a little pasta water.
3) Finish with cheese/herbs and adjust seasoning.$$, 'https://picsum.photos/seed/pasta-garlic-parmesan-pasta/800/500', 631, 28),
('9bc0b6ff-d7ef-5151-b1a1-13c27dc3b4ac'::uuid, 'Creamy Alfredo (Creamy)', 'Pasta', $$1) Boil pasta in salted water until al dente; reserve a splash of pasta water.
2) Build sauce in a pan, then toss pasta with sauce and a little pasta water.
3) Finish with cheese/herbs and adjust seasoning.$$, 'https://picsum.photos/seed/pasta-creamy-alfredo-creamy/800/500', 793, 32),
('12e9d911-c173-5dab-8b0b-12f189a2b5de'::uuid, 'One-Pot Pasta (Zesty)', 'Pasta', $$1) Boil pasta in salted water until al dente; reserve a splash of pasta water.
2) Build sauce in a pan, then toss pasta with sauce and a little pasta water.
3) Finish with cheese/herbs and adjust seasoning.$$, 'https://picsum.photos/seed/pasta-one-pot-pasta-zesty/800/500', 809, 19),
('f7b2162a-c42d-50c5-b2d0-918661f17e44'::uuid, 'Spicy Arrabbiata', 'Pasta', $$1) Boil pasta in salted water until al dente; reserve a splash of pasta water.
2) Build sauce in a pan, then toss pasta with sauce and a little pasta water.
3) Finish with cheese/herbs and adjust seasoning.$$, 'https://picsum.photos/seed/pasta-spicy-arrabbiata/800/500', 761, 55),
('9c3c5294-948c-5aa4-a85e-2eddd9957919'::uuid, 'Lemon Ricotta Pasta', 'Pasta', $$1) Boil pasta in salted water until al dente; reserve a splash of pasta water.
2) Build sauce in a pan, then toss pasta with sauce and a little pasta water.
3) Finish with cheese/herbs and adjust seasoning.$$, 'https://picsum.photos/seed/pasta-lemon-ricotta-pasta/800/500', 537, 49),
('a8b913c4-a147-55fe-a34f-a440058cf6dd'::uuid, 'Pasta Salad (BBQ)', 'Pasta', $$1) Boil pasta in salted water until al dente; reserve a splash of pasta water.
2) Build sauce in a pan, then toss pasta with sauce and a little pasta water.
3) Finish with cheese/herbs and adjust seasoning.$$, 'https://picsum.photos/seed/pasta-pasta-salad-bbq/800/500', 823, 30),
('7faaf8b8-ec40-51a5-a5f5-dfe7eed91a83'::uuid, 'Mac & Cheese (Stovetop)', 'Pasta', $$1) Boil pasta in salted water until al dente; reserve a splash of pasta water.
2) Build sauce in a pan, then toss pasta with sauce and a little pasta water.
3) Finish with cheese/herbs and adjust seasoning.$$, 'https://picsum.photos/seed/pasta-mac--cheese-stovetop/800/500', 533, 44),
('2d2539b2-c2d6-50f0-977d-0e767a5f69ff'::uuid, 'Honey Garlic Pork Chops', 'Pork', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/pork-honey-garlic-pork-chops/800/500', 694, 54),
('19f403ca-20a3-50b9-abea-e64e7bb6017a'::uuid, 'Pulled Pork (Sweet)', 'Pork', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/pork-pulled-pork-sweet/800/500', 827, 108),
('8c61a08b-f5bd-5d7d-97d2-bf45f38d82e3'::uuid, 'Pork Fried Rice', 'Pork', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/pork-pork-fried-rice/800/500', 785, 48),
('1dbd5559-297c-5246-bdcd-3bcb30f28a1a'::uuid, 'Pork Creamy Tacos', 'Pork', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/pork-pork-creamy-tacos/800/500', 850, 61),
('02b74751-ac90-5faa-b6d1-52d88b482b0b'::uuid, 'Pork Tenderloin (Roasted)', 'Pork', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/pork-pork-tenderloin-roasted/800/500', 931, 118),
('efe492bf-bbf3-521d-826d-44a457f39232'::uuid, 'Pork Ramen Bowl', 'Pork', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/pork-pork-ramen-bowl/800/500', 897, 27),
('4848736f-65b2-57ab-9f1b-62d7645832f9'::uuid, 'Sweet & Sour Pork', 'Pork', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/pork-sweet--sour-pork/800/500', 617, 24),
('c79c78a6-8b65-57c6-9ee7-7939ab956613'::uuid, 'Pork Sausage Pasta', 'Pork', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/pork-pork-sausage-pasta/800/500', 912, 60),
('613a8382-1091-565c-86cc-37022bca1946'::uuid, 'Pork and Cabbage Skillet', 'Pork', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/pork-pork-and-cabbage-skillet/800/500', 705, 54),
('110894b7-adea-507a-bf70-80cea261d9b6'::uuid, 'Carnitas Bowl', 'Pork', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/pork-carnitas-bowl/800/500', 533, 47),
('04f148d1-ab13-5dbe-93c9-a3acce81d221'::uuid, 'Sheet Pan Salmon & Sweet', 'Seafood', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/seafood-sheet-pan-salmon--sweet/800/500', 640, 30),
('32cc11a8-c195-513c-8fa4-44e846b56933'::uuid, 'Garlic Butter Shrimp', 'Seafood', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/seafood-garlic-butter-shrimp/800/500', 458, 41),
('f5f314c0-a8d4-50e8-a2d8-997bf26840f8'::uuid, 'Shrimp Tacos (Creamy)', 'Seafood', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/seafood-shrimp-tacos-creamy/800/500', 552, 39),
('c355d867-7a68-58d3-b389-e8d27704251d'::uuid, 'Tuna Salad Wrap', 'Seafood', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/seafood-tuna-salad-wrap/800/500', 423, 26),
('9501d650-e0c7-545c-8a26-bec5b953a829'::uuid, 'Fish Taco Bowls', 'Seafood', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/seafood-fish-taco-bowls/800/500', 421, 25),
('4a2111a3-f682-511b-b4c4-20b3c6c3eb1b'::uuid, 'Salmon Rice Bowl', 'Seafood', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/seafood-salmon-rice-bowl/800/500', 731, 45),
('cf4d26e6-fd86-5209-91ec-e0ff3e7410bd'::uuid, 'Pasta with Shrimp (BBQ)', 'Seafood', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/seafood-pasta-with-shrimp-bbq/800/500', 625, 26),
('e7e682cc-cd15-544c-8b21-b67cfd4d8aa0'::uuid, 'Baked Cod (Lemon)', 'Seafood', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/seafood-baked-cod-lemon/800/500', 732, 37),
('9fd3101f-ed19-5569-835d-f56e708fb0c4'::uuid, 'Seafood Stir-Fry', 'Seafood', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/seafood-seafood-stir-fry/800/500', 648, 35),
('a0f9ac24-7bd9-5b8c-a195-52f25a9176dc'::uuid, 'Sardine Toast (Basil)', 'Seafood', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/seafood-sardine-toast-basil/800/500', 535, 24),
('7478213c-93a6-52f6-ae7b-631ca7c1540a'::uuid, 'Roasted Garlic Potatoes', 'Side', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/side-roasted-garlic-potatoes/800/500', 190, 42),
('f9523376-1410-501a-844e-a8001137f566'::uuid, 'Simple Coleslaw', 'Side', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/side-simple-coleslaw/800/500', 372, 15),
('50a24b51-1906-5cee-8df0-b75d5b5a46ac'::uuid, 'Garlic Green Beans', 'Side', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/side-garlic-green-beans/800/500', 144, 17),
('efd3e5c0-5d3f-56e7-86af-d2e13b54b80f'::uuid, 'Honey Glazed Carrots', 'Side', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/side-honey-glazed-carrots/800/500', 198, 20),
('98469b7e-a5dc-54f1-9147-0df4a34ba840'::uuid, 'Cucumber Salad (Ginger)', 'Side', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/side-cucumber-salad-ginger/800/500', 336, 14),
('6da33140-98b8-5519-87cf-cffd8307c94c'::uuid, 'Corn on the Cob (BBQ)', 'Side', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/side-corn-on-the-cob-bbq/800/500', 317, 34),
('59265790-cb79-532d-b01c-fea147198fa0'::uuid, 'Roasted Broccoli (Maple)', 'Side', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/side-roasted-broccoli-maple/800/500', 425, 39),
('d7f36ed2-05a7-5063-8b50-d57157577820'::uuid, 'Mashed Potatoes (Chili)', 'Side', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/side-mashed-potatoes-chili/800/500', 390, 26),
('6c0cf6fb-1e64-5095-b4b6-ecf16b6a7ed0'::uuid, 'Side Salad (Basil)', 'Side', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/side-side-salad-basil/800/500', 403, 10),
('017ed580-e30a-5e7e-8fb4-33492b92ede8'::uuid, 'Buttered Rice (Parmesan)', 'Side', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/side-buttered-rice-parmesan/800/500', 178, 44),
('d8af8860-c8d5-5807-a91d-e5dc20b176c9'::uuid, 'Chickpea Coconut Curry', 'Vegetarian', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/vegetarian-chickpea-coconut-curry/800/500', 734, 32),
('3e74bf9d-9695-5b01-8aac-7d98058185f2'::uuid, 'Veggie Stir-Fry (Zesty)', 'Vegetarian', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/vegetarian-veggie-stir-fry-zesty/800/500', 743, 56),
('c7684b92-957d-51d1-a0b1-043c9a933a3a'::uuid, 'Hearty Lentil Soup', 'Vegetarian', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/vegetarian-hearty-lentil-soup/800/500', 524, 22),
('e79d4e5e-b4e7-59a9-9954-ba1261ab9522'::uuid, 'Black Bean Bowls (Ginger)', 'Vegetarian', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/vegetarian-black-bean-bowls-ginger/800/500', 500, 42),
('363c757b-b4d6-5f6d-896e-b532accfbc1e'::uuid, 'Tofu BBQ Stir-Fry', 'Vegetarian', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/vegetarian-tofu-bbq-stir-fry/800/500', 430, 44),
('b2b8ba13-ff1e-5c33-8fcf-55af44ac3c2f'::uuid, 'Vegetarian Chili', 'Vegetarian', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/vegetarian-vegetarian-chili/800/500', 351, 61),
('1330a946-51ef-5fb9-8606-f197577377b4'::uuid, 'Roasted Veggie Sheet Pan', 'Vegetarian', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/vegetarian-roasted-veggie-sheet-pan/800/500', 718, 31),
('1a687bb8-fc57-5ac7-b7eb-03bd1800d54c'::uuid, 'Caprese Salad (Basil)', 'Vegetarian', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/vegetarian-caprese-salad-basil/800/500', 606, 63),
('aa29d27e-6068-5ea5-b2ba-5a52ee67a09a'::uuid, 'Veggie Fried Rice', 'Vegetarian', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/vegetarian-veggie-fried-rice/800/500', 441, 47),
('ef3185ab-78e1-5c2e-8ed1-8c9df2016fc2'::uuid, 'Stuffed Peppers (Miso)', 'Vegetarian', $$1) Prep and chop ingredients.
2) Cook protein/veg, then add aromatics and seasonings.
3) Simmer or finish until cooked through.
4) Taste, adjust seasoning, and serve.$$, 'https://picsum.photos/seed/vegetarian-stuffed-peppers-miso/800/500', 404, 70),
('fec99184-5e4b-573d-9c9b-34edb5bb2732'::uuid, 'Fluffy Zesty Pancakes', 'Breakfast', $$1) Prep ingredients and measure everything out.
2) Cook or assemble as directed until set and warmed through.
3) Season lightly and adjust sweetness or salt to taste.
4) Serve immediately or chill if desired.$$, 'https://picsum.photos/seed/breakfast-fluffy-zesty-pancakes/800/500', 570, 14),
('93f134ef-df0b-5279-b22e-e279d5d737b6'::uuid, 'Overnight Oats (Sesame)', 'Breakfast', $$1) Prep ingredients and measure everything out.
2) Cook or assemble as directed until set and warmed through.
3) Season lightly and adjust sweetness or salt to taste.
4) Serve immediately or chill if desired.$$, 'https://picsum.photos/seed/breakfast-overnight-oats-sesame/800/500', 577, 21),
('540b4aad-2c28-5db5-ab30-b2de209d7cc0'::uuid, 'Ginger Breakfast Burrito', 'Breakfast', $$1) Prep ingredients and measure everything out.
2) Cook or assemble as directed until set and warmed through.
3) Season lightly and adjust sweetness or salt to taste.
4) Serve immediately or chill if desired.$$, 'https://picsum.photos/seed/breakfast-ginger-breakfast-burrito/800/500', 561, 11),
('e152b51b-187a-51a7-b891-0c0f0b64fc90'::uuid, 'BBQ Yogurt Parfait', 'Breakfast', $$1) Prep ingredients and measure everything out.
2) Cook or assemble as directed until set and warmed through.
3) Season lightly and adjust sweetness or salt to taste.
4) Serve immediately or chill if desired.$$, 'https://picsum.photos/seed/breakfast-bbq-yogurt-parfait/800/500', 328, 16),
('5546b970-f852-5b1b-96c7-bfdb825a8330'::uuid, 'Maple Avocado Toast', 'Breakfast', $$1) Prep ingredients and measure everything out.
2) Cook or assemble as directed until set and warmed through.
3) Season lightly and adjust sweetness or salt to taste.
4) Serve immediately or chill if desired.$$, 'https://picsum.photos/seed/breakfast-maple-avocado-toast/800/500', 640, 10),
('8926e00c-061b-5fe4-853c-79a529a2ffa8'::uuid, 'Baked Egg Cups (Chili)', 'Breakfast', $$1) Prep ingredients and measure everything out.
2) Cook or assemble as directed until set and warmed through.
3) Season lightly and adjust sweetness or salt to taste.
4) Serve immediately or chill if desired.$$, 'https://picsum.photos/seed/breakfast-baked-egg-cups-chili/800/500', 526, 21),
('f6897b89-3037-5065-8ed0-fe235de96584'::uuid, 'French Toast (Basil)', 'Breakfast', $$1) Prep ingredients and measure everything out.
2) Cook or assemble as directed until set and warmed through.
3) Season lightly and adjust sweetness or salt to taste.
4) Serve immediately or chill if desired.$$, 'https://picsum.photos/seed/breakfast-french-toast-basil/800/500', 250, 24),
('0a127797-69be-5039-938d-9b235208b2f3'::uuid, 'Parmesan Breakfast Sandwich', 'Breakfast', $$1) Prep ingredients and measure everything out.
2) Cook or assemble as directed until set and warmed through.
3) Season lightly and adjust sweetness or salt to taste.
4) Serve immediately or chill if desired.$$, 'https://picsum.photos/seed/breakfast-parmesan-breakfast-sandwich/800/500', 415, 20),
('c3ea2d31-7b34-5032-baed-fa71669cb614'::uuid, 'Miso Granola Bowl', 'Breakfast', $$1) Prep ingredients and measure everything out.
2) Cook or assemble as directed until set and warmed through.
3) Season lightly and adjust sweetness or salt to taste.
4) Serve immediately or chill if desired.$$, 'https://picsum.photos/seed/breakfast-miso-granola-bowl/800/500', 259, 8),
('e5ff2a13-cc76-5dd2-bda1-53ed68ee5fd7'::uuid, 'Coconut Smoothie', 'Breakfast', $$1) Prep ingredients and measure everything out.
2) Cook or assemble as directed until set and warmed through.
3) Season lightly and adjust sweetness or salt to taste.
4) Serve immediately or chill if desired.$$, 'https://picsum.photos/seed/breakfast-coconut-smoothie/800/500', 435, 14);

-- Upsert recipes for the demo user
WITH demo AS (
  SELECT id
  FROM users
  WHERE email = 'example@gmail.com'
  LIMIT 1
)
INSERT INTO recipes (id, user_id, name, category, instructions, thumbnail_url, calories, total_cook_time)
SELECT s.id, demo.id, s.name, s.category, s.instructions, s.thumbnail_url, s.calories, s.total_cook_time
FROM seed_recipes s
CROSS JOIN demo
ON CONFLICT (id) DO UPDATE SET
  user_id         = EXCLUDED.user_id,
  name            = EXCLUDED.name,
  category        = EXCLUDED.category,
  instructions    = EXCLUDED.instructions,
  thumbnail_url   = EXCLUDED.thumbnail_url,
  calories        = EXCLUDED.calories,
  total_cook_time = EXCLUDED.total_cook_time;

-- Replace ingredients for these recipe IDs (prevents duplicates on re-run)
DELETE FROM ingredients WHERE recipe_id IN (SELECT id FROM seed_recipes);

CREATE TEMP TABLE seed_ingredients (
  recipe_id uuid NOT NULL,
  name text NOT NULL,
  measurement text NOT NULL
) ON COMMIT DROP;

INSERT INTO seed_ingredients (recipe_id, name, measurement)
VALUES
('ced3ffc1-d43b-576c-8a7c-5579e64647aa'::uuid, 'Flank steak', '1 lb'),
('ced3ffc1-d43b-576c-8a7c-5579e64647aa'::uuid, 'Rice', '1 cup uncooked'),
('ced3ffc1-d43b-576c-8a7c-5579e64647aa'::uuid, 'Tortillas', '8'),
('ced3ffc1-d43b-576c-8a7c-5579e64647aa'::uuid, 'Olive oil', '2 tbsp'),
('ced3ffc1-d43b-576c-8a7c-5579e64647aa'::uuid, 'Kidney beans', '1 (15 oz) can'),
('ced3ffc1-d43b-576c-8a7c-5579e64647aa'::uuid, 'Salt', 'to taste'),
('ced3ffc1-d43b-576c-8a7c-5579e64647aa'::uuid, 'Black pepper', 'to taste'),
('ced3ffc1-d43b-576c-8a7c-5579e64647aa'::uuid, 'Soy sauce', '3 tbsp'),
('ced3ffc1-d43b-576c-8a7c-5579e64647aa'::uuid, 'Salsa', '1/2 cup'),
('5c584359-337b-5e7d-80c6-dd385f6881ec'::uuid, 'Salsa', '1/2 cup'),
('5c584359-337b-5e7d-80c6-dd385f6881ec'::uuid, 'Salt', 'to taste'),
('5c584359-337b-5e7d-80c6-dd385f6881ec'::uuid, 'Garlic', '3 cloves'),
('5c584359-337b-5e7d-80c6-dd385f6881ec'::uuid, 'Soy sauce', '3 tbsp'),
('5c584359-337b-5e7d-80c6-dd385f6881ec'::uuid, 'Black pepper', 'to taste'),
('5c584359-337b-5e7d-80c6-dd385f6881ec'::uuid, 'Cheddar cheese', '1 cup'),
('5c584359-337b-5e7d-80c6-dd385f6881ec'::uuid, 'Chili powder', '2 tbsp'),
('5c584359-337b-5e7d-80c6-dd385f6881ec'::uuid, 'Bell pepper', '1'),
('5c584359-337b-5e7d-80c6-dd385f6881ec'::uuid, 'Taco seasoning', '2 tbsp'),
('2cfeec0f-31a7-5b27-bf77-8e5e6563da84'::uuid, 'Bell pepper', '1'),
('2cfeec0f-31a7-5b27-bf77-8e5e6563da84'::uuid, 'Chili powder', '2 tbsp'),
('2cfeec0f-31a7-5b27-bf77-8e5e6563da84'::uuid, 'Crushed tomatoes', '1 (28 oz) can'),
('2cfeec0f-31a7-5b27-bf77-8e5e6563da84'::uuid, 'Kidney beans', '1 (15 oz) can'),
('2cfeec0f-31a7-5b27-bf77-8e5e6563da84'::uuid, 'Rice', '1 cup uncooked'),
('2cfeec0f-31a7-5b27-bf77-8e5e6563da84'::uuid, 'Soy sauce', '3 tbsp'),
('2cfeec0f-31a7-5b27-bf77-8e5e6563da84'::uuid, 'Lettuce', '2 cups'),
('2cfeec0f-31a7-5b27-bf77-8e5e6563da84'::uuid, 'Flank steak', '1 lb'),
('2cfeec0f-31a7-5b27-bf77-8e5e6563da84'::uuid, 'Taco seasoning', '2 tbsp'),
('2cfeec0f-31a7-5b27-bf77-8e5e6563da84'::uuid, 'Cumin', '1 tbsp'),
('2cfeec0f-31a7-5b27-bf77-8e5e6563da84'::uuid, 'Salsa', '1/2 cup'),
('2cfeec0f-31a7-5b27-bf77-8e5e6563da84'::uuid, 'Salt', 'to taste'),
('e410dd2f-f321-564c-9cec-a3cfb036c9c0'::uuid, 'Black pepper', 'to taste'),
('e410dd2f-f321-564c-9cec-a3cfb036c9c0'::uuid, 'Salt', 'to taste'),
('e410dd2f-f321-564c-9cec-a3cfb036c9c0'::uuid, 'Rice', '1 cup uncooked'),
('e410dd2f-f321-564c-9cec-a3cfb036c9c0'::uuid, 'Olive oil', '2 tbsp'),
('e410dd2f-f321-564c-9cec-a3cfb036c9c0'::uuid, 'Ground beef', '1 lb'),
('e410dd2f-f321-564c-9cec-a3cfb036c9c0'::uuid, 'Flank steak', '1 lb'),
('e410dd2f-f321-564c-9cec-a3cfb036c9c0'::uuid, 'Kidney beans', '1 (15 oz) can'),
('e410dd2f-f321-564c-9cec-a3cfb036c9c0'::uuid, 'Cumin', '1 tbsp'),
('133b2dc5-06f4-5ff3-9eef-c3a9370c0680'::uuid, 'Rice', '1 cup uncooked'),
('133b2dc5-06f4-5ff3-9eef-c3a9370c0680'::uuid, 'Tortillas', '8'),
('133b2dc5-06f4-5ff3-9eef-c3a9370c0680'::uuid, 'Flank steak', '1 lb'),
('133b2dc5-06f4-5ff3-9eef-c3a9370c0680'::uuid, 'Cumin', '1 tbsp'),
('133b2dc5-06f4-5ff3-9eef-c3a9370c0680'::uuid, 'Olive oil', '2 tbsp'),
('133b2dc5-06f4-5ff3-9eef-c3a9370c0680'::uuid, 'Soy sauce', '3 tbsp'),
('133b2dc5-06f4-5ff3-9eef-c3a9370c0680'::uuid, 'Taco seasoning', '2 tbsp'),
('133b2dc5-06f4-5ff3-9eef-c3a9370c0680'::uuid, 'Salt', 'to taste'),
('ef51d7ce-9ecd-50fe-b653-c951faf008e8'::uuid, 'Bell pepper', '1'),
('ef51d7ce-9ecd-50fe-b653-c951faf008e8'::uuid, 'Salt', 'to taste'),
('ef51d7ce-9ecd-50fe-b653-c951faf008e8'::uuid, 'Onion', '1'),
('ef51d7ce-9ecd-50fe-b653-c951faf008e8'::uuid, 'Salsa', '1/2 cup'),
('ef51d7ce-9ecd-50fe-b653-c951faf008e8'::uuid, 'Taco seasoning', '2 tbsp'),
('ef51d7ce-9ecd-50fe-b653-c951faf008e8'::uuid, 'Crushed tomatoes', '1 (28 oz) can'),
('ef51d7ce-9ecd-50fe-b653-c951faf008e8'::uuid, 'Rice', '1 cup uncooked'),
('ef51d7ce-9ecd-50fe-b653-c951faf008e8'::uuid, 'Black pepper', 'to taste'),
('ef51d7ce-9ecd-50fe-b653-c951faf008e8'::uuid, 'Lettuce', '2 cups'),
('ef51d7ce-9ecd-50fe-b653-c951faf008e8'::uuid, 'Flank steak', '1 lb'),
('ef51d7ce-9ecd-50fe-b653-c951faf008e8'::uuid, 'Chili powder', '2 tbsp'),
('cc80dc63-8e54-5cf7-821c-97961e58f84e'::uuid, 'Kidney beans', '1 (15 oz) can'),
('cc80dc63-8e54-5cf7-821c-97961e58f84e'::uuid, 'Cheddar cheese', '1 cup'),
('cc80dc63-8e54-5cf7-821c-97961e58f84e'::uuid, 'Salt', 'to taste'),
('cc80dc63-8e54-5cf7-821c-97961e58f84e'::uuid, 'Lettuce', '2 cups'),
('cc80dc63-8e54-5cf7-821c-97961e58f84e'::uuid, 'Olive oil', '2 tbsp'),
('cc80dc63-8e54-5cf7-821c-97961e58f84e'::uuid, 'Black pepper', 'to taste'),
('cc80dc63-8e54-5cf7-821c-97961e58f84e'::uuid, 'Ground beef', '1 lb'),
('cc80dc63-8e54-5cf7-821c-97961e58f84e'::uuid, 'Cumin', '1 tbsp'),
('cc80dc63-8e54-5cf7-821c-97961e58f84e'::uuid, 'Salsa', '1/2 cup'),
('cc80dc63-8e54-5cf7-821c-97961e58f84e'::uuid, 'Flank steak', '1 lb'),
('cc80dc63-8e54-5cf7-821c-97961e58f84e'::uuid, 'Crushed tomatoes', '1 (28 oz) can'),
('9cdbc9bb-02c3-552d-9e96-36036f40e948'::uuid, 'Cumin', '1 tbsp'),
('9cdbc9bb-02c3-552d-9e96-36036f40e948'::uuid, 'Taco seasoning', '2 tbsp'),
('9cdbc9bb-02c3-552d-9e96-36036f40e948'::uuid, 'Rice', '1 cup uncooked'),
('9cdbc9bb-02c3-552d-9e96-36036f40e948'::uuid, 'Bell pepper', '1'),
('9cdbc9bb-02c3-552d-9e96-36036f40e948'::uuid, 'Salt', 'to taste'),
('9cdbc9bb-02c3-552d-9e96-36036f40e948'::uuid, 'Soy sauce', '3 tbsp'),
('9cdbc9bb-02c3-552d-9e96-36036f40e948'::uuid, 'Olive oil', '2 tbsp'),
('9cdbc9bb-02c3-552d-9e96-36036f40e948'::uuid, 'Tortillas', '8'),
('9cdbc9bb-02c3-552d-9e96-36036f40e948'::uuid, 'Salsa', '1/2 cup'),
('9cdbc9bb-02c3-552d-9e96-36036f40e948'::uuid, 'Kidney beans', '1 (15 oz) can'),
('9cdbc9bb-02c3-552d-9e96-36036f40e948'::uuid, 'Onion', '1'),
('2866b2a1-3eb7-506b-9065-ae7c861d5de1'::uuid, 'Rice', '1 cup uncooked'),
('2866b2a1-3eb7-506b-9065-ae7c861d5de1'::uuid, 'Tortillas', '8'),
('2866b2a1-3eb7-506b-9065-ae7c861d5de1'::uuid, 'Lettuce', '2 cups'),
('2866b2a1-3eb7-506b-9065-ae7c861d5de1'::uuid, 'Taco seasoning', '2 tbsp'),
('2866b2a1-3eb7-506b-9065-ae7c861d5de1'::uuid, 'Ground beef', '1 lb'),
('2866b2a1-3eb7-506b-9065-ae7c861d5de1'::uuid, 'Cumin', '1 tbsp'),
('2866b2a1-3eb7-506b-9065-ae7c861d5de1'::uuid, 'Soy sauce', '3 tbsp'),
('2866b2a1-3eb7-506b-9065-ae7c861d5de1'::uuid, 'Olive oil', '2 tbsp'),
('2866b2a1-3eb7-506b-9065-ae7c861d5de1'::uuid, 'Flank steak', '1 lb'),
('2866b2a1-3eb7-506b-9065-ae7c861d5de1'::uuid, 'Salsa', '1/2 cup'),
('2866b2a1-3eb7-506b-9065-ae7c861d5de1'::uuid, 'Salt', 'to taste'),
('558ac48b-627e-5b8e-afb6-a88d05cbda69'::uuid, 'Salsa', '1/2 cup'),
('558ac48b-627e-5b8e-afb6-a88d05cbda69'::uuid, 'Black pepper', 'to taste'),
('558ac48b-627e-5b8e-afb6-a88d05cbda69'::uuid, 'Bell pepper', '1'),
('558ac48b-627e-5b8e-afb6-a88d05cbda69'::uuid, 'Crushed tomatoes', '1 (28 oz) can'),
('558ac48b-627e-5b8e-afb6-a88d05cbda69'::uuid, 'Lettuce', '2 cups'),
('558ac48b-627e-5b8e-afb6-a88d05cbda69'::uuid, 'Ground beef', '1 lb'),
('558ac48b-627e-5b8e-afb6-a88d05cbda69'::uuid, 'Tortillas', '8'),
('558ac48b-627e-5b8e-afb6-a88d05cbda69'::uuid, 'Olive oil', '2 tbsp'),
('558ac48b-627e-5b8e-afb6-a88d05cbda69'::uuid, 'Cheddar cheese', '1 cup'),
('558ac48b-627e-5b8e-afb6-a88d05cbda69'::uuid, 'Kidney beans', '1 (15 oz) can'),
('558ac48b-627e-5b8e-afb6-a88d05cbda69'::uuid, 'Salt', 'to taste'),
('6e70f710-b1e7-5747-86d7-85cd61e9d648'::uuid, 'Rice', '1 cup uncooked'),
('6e70f710-b1e7-5747-86d7-85cd61e9d648'::uuid, 'Tortillas', '8'),
('6e70f710-b1e7-5747-86d7-85cd61e9d648'::uuid, 'Tomato sauce', '2 cups'),
('6e70f710-b1e7-5747-86d7-85cd61e9d648'::uuid, 'Olive oil', '2 tbsp'),
('6e70f710-b1e7-5747-86d7-85cd61e9d648'::uuid, 'Lemon', '1'),
('6e70f710-b1e7-5747-86d7-85cd61e9d648'::uuid, 'Garlic', '3 cloves'),
('6e70f710-b1e7-5747-86d7-85cd61e9d648'::uuid, 'Salt', 'to taste'),
('6e70f710-b1e7-5747-86d7-85cd61e9d648'::uuid, 'Chicken breast', '1 lb'),
('6e70f710-b1e7-5747-86d7-85cd61e9d648'::uuid, 'Black pepper', 'to taste'),
('6e70f710-b1e7-5747-86d7-85cd61e9d648'::uuid, 'Onion', '1'),
('6e70f710-b1e7-5747-86d7-85cd61e9d648'::uuid, 'Broccoli', '3 cups'),
('e22aaba5-57ed-5a58-83f9-bd3ec6c4c178'::uuid, 'Chicken thighs', '1 lb'),
('e22aaba5-57ed-5a58-83f9-bd3ec6c4c178'::uuid, 'Black pepper', 'to taste'),
('e22aaba5-57ed-5a58-83f9-bd3ec6c4c178'::uuid, 'Tomato sauce', '2 cups'),
('e22aaba5-57ed-5a58-83f9-bd3ec6c4c178'::uuid, 'Onion', '1'),
('e22aaba5-57ed-5a58-83f9-bd3ec6c4c178'::uuid, 'Tortillas', '8'),
('e22aaba5-57ed-5a58-83f9-bd3ec6c4c178'::uuid, 'Parmesan', '1/2 cup'),
('e22aaba5-57ed-5a58-83f9-bd3ec6c4c178'::uuid, 'Olive oil', '2 tbsp'),
('e22aaba5-57ed-5a58-83f9-bd3ec6c4c178'::uuid, 'Chicken breast', '1 lb'),
('e22aaba5-57ed-5a58-83f9-bd3ec6c4c178'::uuid, 'Chicken broth', '3 cups'),
('e22aaba5-57ed-5a58-83f9-bd3ec6c4c178'::uuid, 'Salt', 'to taste'),
('9d7b95e2-2923-5160-aff6-03de0f3dc88b'::uuid, 'Olive oil', '2 tbsp'),
('9d7b95e2-2923-5160-aff6-03de0f3dc88b'::uuid, 'Black pepper', 'to taste'),
('9d7b95e2-2923-5160-aff6-03de0f3dc88b'::uuid, 'Bell pepper', '1'),
('9d7b95e2-2923-5160-aff6-03de0f3dc88b'::uuid, 'Honey', '1 tbsp'),
('9d7b95e2-2923-5160-aff6-03de0f3dc88b'::uuid, 'Chicken thighs', '1 lb'),
('9d7b95e2-2923-5160-aff6-03de0f3dc88b'::uuid, 'Rice', '1 cup uncooked'),
('9d7b95e2-2923-5160-aff6-03de0f3dc88b'::uuid, 'Lemon', '1'),
('9d7b95e2-2923-5160-aff6-03de0f3dc88b'::uuid, 'Chicken broth', '3 cups'),
('9d7b95e2-2923-5160-aff6-03de0f3dc88b'::uuid, 'Salt', 'to taste'),
('634d8556-441d-5180-90a4-347a256e4540'::uuid, 'Olive oil', '2 tbsp'),
('634d8556-441d-5180-90a4-347a256e4540'::uuid, 'Tortillas', '8'),
('634d8556-441d-5180-90a4-347a256e4540'::uuid, 'Broccoli', '3 cups'),
('634d8556-441d-5180-90a4-347a256e4540'::uuid, 'Parmesan', '1/2 cup'),
('634d8556-441d-5180-90a4-347a256e4540'::uuid, 'Garlic', '3 cloves'),
('634d8556-441d-5180-90a4-347a256e4540'::uuid, 'Lemon', '1'),
('634d8556-441d-5180-90a4-347a256e4540'::uuid, 'Tomato sauce', '2 cups'),
('634d8556-441d-5180-90a4-347a256e4540'::uuid, 'Salt', 'to taste'),
('49272eb3-1b6c-5cd3-859d-5c0a6e2bda4f'::uuid, 'Onion', '1'),
('49272eb3-1b6c-5cd3-859d-5c0a6e2bda4f'::uuid, 'Chicken broth', '3 cups'),
('49272eb3-1b6c-5cd3-859d-5c0a6e2bda4f'::uuid, 'Garlic', '3 cloves'),
('49272eb3-1b6c-5cd3-859d-5c0a6e2bda4f'::uuid, 'Broccoli', '3 cups'),
('49272eb3-1b6c-5cd3-859d-5c0a6e2bda4f'::uuid, 'Parmesan', '1/2 cup'),
('49272eb3-1b6c-5cd3-859d-5c0a6e2bda4f'::uuid, 'Tomato sauce', '2 cups'),
('49272eb3-1b6c-5cd3-859d-5c0a6e2bda4f'::uuid, 'Bell pepper', '1'),
('49272eb3-1b6c-5cd3-859d-5c0a6e2bda4f'::uuid, 'Soy sauce', '3 tbsp'),
('49272eb3-1b6c-5cd3-859d-5c0a6e2bda4f'::uuid, 'Salt', 'to taste'),
('a63fbab2-fa52-5a49-a507-ef7f9f9e62cf'::uuid, 'Parmesan', '1/2 cup'),
('a63fbab2-fa52-5a49-a507-ef7f9f9e62cf'::uuid, 'Lemon', '1'),
('a63fbab2-fa52-5a49-a507-ef7f9f9e62cf'::uuid, 'Olive oil', '2 tbsp'),
('a63fbab2-fa52-5a49-a507-ef7f9f9e62cf'::uuid, 'Onion', '1'),
('a63fbab2-fa52-5a49-a507-ef7f9f9e62cf'::uuid, 'Salt', 'to taste'),
('a63fbab2-fa52-5a49-a507-ef7f9f9e62cf'::uuid, 'Black pepper', 'to taste'),
('a63fbab2-fa52-5a49-a507-ef7f9f9e62cf'::uuid, 'Garlic', '3 cloves'),
('a63fbab2-fa52-5a49-a507-ef7f9f9e62cf'::uuid, 'Tomato sauce', '2 cups'),
('19aeb489-becf-5485-adc7-f6d12917032c'::uuid, 'Olive oil', '2 tbsp'),
('19aeb489-becf-5485-adc7-f6d12917032c'::uuid, 'Bell pepper', '1'),
('19aeb489-becf-5485-adc7-f6d12917032c'::uuid, 'Teriyaki sauce', '1/3 cup'),
('19aeb489-becf-5485-adc7-f6d12917032c'::uuid, 'Garlic', '3 cloves'),
('19aeb489-becf-5485-adc7-f6d12917032c'::uuid, 'Black pepper', 'to taste'),
('19aeb489-becf-5485-adc7-f6d12917032c'::uuid, 'Salt', 'to taste'),
('19aeb489-becf-5485-adc7-f6d12917032c'::uuid, 'Onion', '1'),
('19aeb489-becf-5485-adc7-f6d12917032c'::uuid, 'Parmesan', '1/2 cup'),
('19aeb489-becf-5485-adc7-f6d12917032c'::uuid, 'Rice', '1 cup uncooked'),
('19aeb489-becf-5485-adc7-f6d12917032c'::uuid, 'Tomato sauce', '2 cups'),
('e1b6903b-3107-58e3-a305-b17355f68c25'::uuid, 'Black pepper', 'to taste'),
('e1b6903b-3107-58e3-a305-b17355f68c25'::uuid, 'Rice', '1 cup uncooked'),
('e1b6903b-3107-58e3-a305-b17355f68c25'::uuid, 'Lemon', '1'),
('e1b6903b-3107-58e3-a305-b17355f68c25'::uuid, 'Parmesan', '1/2 cup'),
('e1b6903b-3107-58e3-a305-b17355f68c25'::uuid, 'Olive oil', '2 tbsp'),
('e1b6903b-3107-58e3-a305-b17355f68c25'::uuid, 'Garlic', '3 cloves'),
('e1b6903b-3107-58e3-a305-b17355f68c25'::uuid, 'Chicken thighs', '1 lb'),
('e1b6903b-3107-58e3-a305-b17355f68c25'::uuid, 'Salt', 'to taste'),
('d80397f2-d891-5bff-b189-a4f8746ac62d'::uuid, 'Garlic', '3 cloves'),
('d80397f2-d891-5bff-b189-a4f8746ac62d'::uuid, 'Onion', '1'),
('d80397f2-d891-5bff-b189-a4f8746ac62d'::uuid, 'Rice', '1 cup uncooked'),
('d80397f2-d891-5bff-b189-a4f8746ac62d'::uuid, 'Tomato sauce', '2 cups'),
('d80397f2-d891-5bff-b189-a4f8746ac62d'::uuid, 'Lemon', '1'),
('d80397f2-d891-5bff-b189-a4f8746ac62d'::uuid, 'Teriyaki sauce', '1/3 cup'),
('d80397f2-d891-5bff-b189-a4f8746ac62d'::uuid, 'Soy sauce', '3 tbsp'),
('d80397f2-d891-5bff-b189-a4f8746ac62d'::uuid, 'Tortillas', '8'),
('d80397f2-d891-5bff-b189-a4f8746ac62d'::uuid, 'Black pepper', 'to taste'),
('d80397f2-d891-5bff-b189-a4f8746ac62d'::uuid, 'Salt', 'to taste'),
('d80397f2-d891-5bff-b189-a4f8746ac62d'::uuid, 'Bell pepper', '1'),
('d80397f2-d891-5bff-b189-a4f8746ac62d'::uuid, 'Olive oil', '2 tbsp'),
('ada2dc0d-afd7-5613-8d26-fba3697b5ceb'::uuid, 'Olive oil', '2 tbsp'),
('ada2dc0d-afd7-5613-8d26-fba3697b5ceb'::uuid, 'Tortillas', '8'),
('ada2dc0d-afd7-5613-8d26-fba3697b5ceb'::uuid, 'Onion', '1'),
('ada2dc0d-afd7-5613-8d26-fba3697b5ceb'::uuid, 'Chicken breast', '1 lb'),
('ada2dc0d-afd7-5613-8d26-fba3697b5ceb'::uuid, 'Parmesan', '1/2 cup'),
('ada2dc0d-afd7-5613-8d26-fba3697b5ceb'::uuid, 'Soy sauce', '3 tbsp'),
('ada2dc0d-afd7-5613-8d26-fba3697b5ceb'::uuid, 'Black pepper', 'to taste'),
('ada2dc0d-afd7-5613-8d26-fba3697b5ceb'::uuid, 'Salt', 'to taste'),
('0bc39706-d1d5-57fa-b5bc-d2fcd21f75ef'::uuid, 'Butter', '1/2 cup'),
('0bc39706-d1d5-57fa-b5bc-d2fcd21f75ef'::uuid, 'Brown sugar', '1/2 cup'),
('0bc39706-d1d5-57fa-b5bc-d2fcd21f75ef'::uuid, 'Cocoa powder', '1/2 cup'),
('0bc39706-d1d5-57fa-b5bc-d2fcd21f75ef'::uuid, 'Chocolate chips', '1 cup'),
('0bc39706-d1d5-57fa-b5bc-d2fcd21f75ef'::uuid, 'Baking soda', '1 tsp'),
('0bc39706-d1d5-57fa-b5bc-d2fcd21f75ef'::uuid, 'Cream cheese', '8 oz'),
('0bc39706-d1d5-57fa-b5bc-d2fcd21f75ef'::uuid, 'Flour', '2 cups'),
('0bc39706-d1d5-57fa-b5bc-d2fcd21f75ef'::uuid, 'Sugar', '1 cup'),
('0bc39706-d1d5-57fa-b5bc-d2fcd21f75ef'::uuid, 'Bananas', '3'),
('0bc39706-d1d5-57fa-b5bc-d2fcd21f75ef'::uuid, 'Salt', '1/2 tsp'),
('7c2e30ef-cede-58f6-b238-10adbbf4c0be'::uuid, 'Chocolate chips', '1 cup'),
('7c2e30ef-cede-58f6-b238-10adbbf4c0be'::uuid, 'Flour', '2 cups'),
('7c2e30ef-cede-58f6-b238-10adbbf4c0be'::uuid, 'Salt', '1/2 tsp'),
('7c2e30ef-cede-58f6-b238-10adbbf4c0be'::uuid, 'Bananas', '3'),
('7c2e30ef-cede-58f6-b238-10adbbf4c0be'::uuid, 'Cinnamon', '1 tsp'),
('7c2e30ef-cede-58f6-b238-10adbbf4c0be'::uuid, 'Butter', '1/2 cup'),
('7c2e30ef-cede-58f6-b238-10adbbf4c0be'::uuid, 'Baking soda', '1 tsp'),
('7c2e30ef-cede-58f6-b238-10adbbf4c0be'::uuid, 'Brown sugar', '1/2 cup'),
('7c2e30ef-cede-58f6-b238-10adbbf4c0be'::uuid, 'Cream cheese', '8 oz'),
('98746583-0d9f-5278-a14c-fc7b545ab8b8'::uuid, 'Salt', '1/2 tsp'),
('98746583-0d9f-5278-a14c-fc7b545ab8b8'::uuid, 'Lemon', '1'),
('98746583-0d9f-5278-a14c-fc7b545ab8b8'::uuid, 'Flour', '2 cups'),
('98746583-0d9f-5278-a14c-fc7b545ab8b8'::uuid, 'Cinnamon', '1 tsp'),
('98746583-0d9f-5278-a14c-fc7b545ab8b8'::uuid, 'Eggs', '2'),
('98746583-0d9f-5278-a14c-fc7b545ab8b8'::uuid, 'Oats', '1 cup'),
('98746583-0d9f-5278-a14c-fc7b545ab8b8'::uuid, 'Sugar', '1 cup'),
('98746583-0d9f-5278-a14c-fc7b545ab8b8'::uuid, 'Brown sugar', '1/2 cup'),
('98746583-0d9f-5278-a14c-fc7b545ab8b8'::uuid, 'Baking soda', '1 tsp'),
('98746583-0d9f-5278-a14c-fc7b545ab8b8'::uuid, 'Vanilla extract', '1 tsp'),
('c0b3a766-aa45-5938-95cd-a747d75afc78'::uuid, 'Bananas', '3'),
('c0b3a766-aa45-5938-95cd-a747d75afc78'::uuid, 'Brown sugar', '1/2 cup'),
('c0b3a766-aa45-5938-95cd-a747d75afc78'::uuid, 'Butter', '1/2 cup'),
('c0b3a766-aa45-5938-95cd-a747d75afc78'::uuid, 'Eggs', '2'),
('c0b3a766-aa45-5938-95cd-a747d75afc78'::uuid, 'Lemon', '1'),
('c0b3a766-aa45-5938-95cd-a747d75afc78'::uuid, 'Oats', '1 cup'),
('c0b3a766-aa45-5938-95cd-a747d75afc78'::uuid, 'Baking soda', '1 tsp'),
('c0b3a766-aa45-5938-95cd-a747d75afc78'::uuid, 'Flour', '2 cups'),
('c0b3a766-aa45-5938-95cd-a747d75afc78'::uuid, 'Sugar', '1 cup'),
('c0b3a766-aa45-5938-95cd-a747d75afc78'::uuid, 'Salt', '1/2 tsp'),
('c0b3a766-aa45-5938-95cd-a747d75afc78'::uuid, 'Cinnamon', '1 tsp'),
('55f27fee-fbb9-5a13-8c36-f122413f5fcc'::uuid, 'Lemon', '1'),
('55f27fee-fbb9-5a13-8c36-f122413f5fcc'::uuid, 'Oats', '1 cup'),
('55f27fee-fbb9-5a13-8c36-f122413f5fcc'::uuid, 'Brown sugar', '1/2 cup'),
('55f27fee-fbb9-5a13-8c36-f122413f5fcc'::uuid, 'Eggs', '2'),
('55f27fee-fbb9-5a13-8c36-f122413f5fcc'::uuid, 'Vanilla extract', '1 tsp'),
('55f27fee-fbb9-5a13-8c36-f122413f5fcc'::uuid, 'Butter', '1/2 cup'),
('55f27fee-fbb9-5a13-8c36-f122413f5fcc'::uuid, 'Sugar', '1 cup'),
('55f27fee-fbb9-5a13-8c36-f122413f5fcc'::uuid, 'Baking soda', '1 tsp'),
('55f27fee-fbb9-5a13-8c36-f122413f5fcc'::uuid, 'Cocoa powder', '1/2 cup'),
('55f27fee-fbb9-5a13-8c36-f122413f5fcc'::uuid, 'Flour', '2 cups'),
('55f27fee-fbb9-5a13-8c36-f122413f5fcc'::uuid, 'Cinnamon', '1 tsp'),
('8d90b5b8-9a45-5623-9006-5b2e4b73dd68'::uuid, 'Eggs', '2'),
('8d90b5b8-9a45-5623-9006-5b2e4b73dd68'::uuid, 'Cream cheese', '8 oz'),
('8d90b5b8-9a45-5623-9006-5b2e4b73dd68'::uuid, 'Cocoa powder', '1/2 cup'),
('8d90b5b8-9a45-5623-9006-5b2e4b73dd68'::uuid, 'Salt', '1/2 tsp'),
('8d90b5b8-9a45-5623-9006-5b2e4b73dd68'::uuid, 'Vanilla extract', '1 tsp'),
('8d90b5b8-9a45-5623-9006-5b2e4b73dd68'::uuid, 'Cinnamon', '1 tsp'),
('8d90b5b8-9a45-5623-9006-5b2e4b73dd68'::uuid, 'Bananas', '3'),
('8d90b5b8-9a45-5623-9006-5b2e4b73dd68'::uuid, 'Flour', '2 cups'),
('8d90b5b8-9a45-5623-9006-5b2e4b73dd68'::uuid, 'Brown sugar', '1/2 cup'),
('3bd1c196-f1c9-580f-bd62-42e17cf162da'::uuid, 'Baking soda', '1 tsp'),
('3bd1c196-f1c9-580f-bd62-42e17cf162da'::uuid, 'Salt', '1/2 tsp'),
('3bd1c196-f1c9-580f-bd62-42e17cf162da'::uuid, 'Vanilla extract', '1 tsp'),
('3bd1c196-f1c9-580f-bd62-42e17cf162da'::uuid, 'Sugar', '1 cup'),
('3bd1c196-f1c9-580f-bd62-42e17cf162da'::uuid, 'Lemon', '1'),
('3bd1c196-f1c9-580f-bd62-42e17cf162da'::uuid, 'Cream cheese', '8 oz'),
('3bd1c196-f1c9-580f-bd62-42e17cf162da'::uuid, 'Chocolate chips', '1 cup'),
('3bd1c196-f1c9-580f-bd62-42e17cf162da'::uuid, 'Cinnamon', '1 tsp'),
('3bd1c196-f1c9-580f-bd62-42e17cf162da'::uuid, 'Bananas', '3'),
('c4e30e25-3b9e-518c-a699-5d02cec395ae'::uuid, 'Salt', '1/2 tsp'),
('c4e30e25-3b9e-518c-a699-5d02cec395ae'::uuid, 'Flour', '2 cups'),
('c4e30e25-3b9e-518c-a699-5d02cec395ae'::uuid, 'Sugar', '1 cup'),
('c4e30e25-3b9e-518c-a699-5d02cec395ae'::uuid, 'Vanilla extract', '1 tsp'),
('c4e30e25-3b9e-518c-a699-5d02cec395ae'::uuid, 'Butter', '1/2 cup'),
('c4e30e25-3b9e-518c-a699-5d02cec395ae'::uuid, 'Bananas', '3'),
('c4e30e25-3b9e-518c-a699-5d02cec395ae'::uuid, 'Brown sugar', '1/2 cup'),
('c4e30e25-3b9e-518c-a699-5d02cec395ae'::uuid, 'Cream cheese', '8 oz'),
('c4e30e25-3b9e-518c-a699-5d02cec395ae'::uuid, 'Cocoa powder', '1/2 cup'),
('c4e30e25-3b9e-518c-a699-5d02cec395ae'::uuid, 'Chocolate chips', '1 cup'),
('c4e30e25-3b9e-518c-a699-5d02cec395ae'::uuid, 'Eggs', '2'),
('c4e30e25-3b9e-518c-a699-5d02cec395ae'::uuid, 'Oats', '1 cup'),
('066df4b8-d7d8-5cb4-9438-4f72c9c903bb'::uuid, 'Baking soda', '1 tsp'),
('066df4b8-d7d8-5cb4-9438-4f72c9c903bb'::uuid, 'Bananas', '3'),
('066df4b8-d7d8-5cb4-9438-4f72c9c903bb'::uuid, 'Chocolate chips', '1 cup'),
('066df4b8-d7d8-5cb4-9438-4f72c9c903bb'::uuid, 'Sugar', '1 cup'),
('066df4b8-d7d8-5cb4-9438-4f72c9c903bb'::uuid, 'Cinnamon', '1 tsp'),
('066df4b8-d7d8-5cb4-9438-4f72c9c903bb'::uuid, 'Cream cheese', '8 oz'),
('066df4b8-d7d8-5cb4-9438-4f72c9c903bb'::uuid, 'Eggs', '2'),
('066df4b8-d7d8-5cb4-9438-4f72c9c903bb'::uuid, 'Vanilla extract', '1 tsp'),
('066df4b8-d7d8-5cb4-9438-4f72c9c903bb'::uuid, 'Flour', '2 cups'),
('066df4b8-d7d8-5cb4-9438-4f72c9c903bb'::uuid, 'Salt', '1/2 tsp'),
('bcd395fe-4f1e-52ab-8c25-9e16dc22020a'::uuid, 'Flour', '2 cups'),
('bcd395fe-4f1e-52ab-8c25-9e16dc22020a'::uuid, 'Chocolate chips', '1 cup'),
('bcd395fe-4f1e-52ab-8c25-9e16dc22020a'::uuid, 'Lemon', '1'),
('bcd395fe-4f1e-52ab-8c25-9e16dc22020a'::uuid, 'Cream cheese', '8 oz'),
('bcd395fe-4f1e-52ab-8c25-9e16dc22020a'::uuid, 'Oats', '1 cup'),
('bcd395fe-4f1e-52ab-8c25-9e16dc22020a'::uuid, 'Eggs', '2'),
('bcd395fe-4f1e-52ab-8c25-9e16dc22020a'::uuid, 'Salt', '1/2 tsp'),
('bcd395fe-4f1e-52ab-8c25-9e16dc22020a'::uuid, 'Baking soda', '1 tsp'),
('bcd395fe-4f1e-52ab-8c25-9e16dc22020a'::uuid, 'Cinnamon', '1 tsp'),
('bcd395fe-4f1e-52ab-8c25-9e16dc22020a'::uuid, 'Brown sugar', '1/2 cup'),
('bcd395fe-4f1e-52ab-8c25-9e16dc22020a'::uuid, 'Butter', '1/2 cup'),
('2981d71a-8174-569f-beaf-f8e601792726'::uuid, 'Canned tomatoes', '2 (14.5 oz) cans'),
('2981d71a-8174-569f-beaf-f8e601792726'::uuid, 'Cilantro', '1/4 cup'),
('2981d71a-8174-569f-beaf-f8e601792726'::uuid, 'Onion', '1'),
('2981d71a-8174-569f-beaf-f8e601792726'::uuid, 'Vegetable broth', '3 cups'),
('2981d71a-8174-569f-beaf-f8e601792726'::uuid, 'Quinoa', '1 cup'),
('2981d71a-8174-569f-beaf-f8e601792726'::uuid, 'Lime', '1'),
('2981d71a-8174-569f-beaf-f8e601792726'::uuid, 'Black pepper', 'to taste'),
('2981d71a-8174-569f-beaf-f8e601792726'::uuid, 'Olive oil', '2 tbsp'),
('2981d71a-8174-569f-beaf-f8e601792726'::uuid, 'Garlic', '2 cloves'),
('2981d71a-8174-569f-beaf-f8e601792726'::uuid, 'Cheddar cheese', '1 cup'),
('2981d71a-8174-569f-beaf-f8e601792726'::uuid, 'Bread', '4 slices'),
('2981d71a-8174-569f-beaf-f8e601792726'::uuid, 'Salt', 'to taste'),
('9a6490b0-e2e5-5aea-bc00-46d3a216e5c5'::uuid, 'Garlic', '2 cloves'),
('9a6490b0-e2e5-5aea-bc00-46d3a216e5c5'::uuid, 'Olive oil', '2 tbsp'),
('9a6490b0-e2e5-5aea-bc00-46d3a216e5c5'::uuid, 'Cheddar cheese', '1 cup'),
('9a6490b0-e2e5-5aea-bc00-46d3a216e5c5'::uuid, 'Cilantro', '1/4 cup'),
('9a6490b0-e2e5-5aea-bc00-46d3a216e5c5'::uuid, 'Black pepper', 'to taste'),
('9a6490b0-e2e5-5aea-bc00-46d3a216e5c5'::uuid, 'Vegetable broth', '3 cups'),
('9a6490b0-e2e5-5aea-bc00-46d3a216e5c5'::uuid, 'Salt', 'to taste'),
('9a6490b0-e2e5-5aea-bc00-46d3a216e5c5'::uuid, 'Bread', '4 slices'),
('9a6490b0-e2e5-5aea-bc00-46d3a216e5c5'::uuid, 'Rice', '1 cup uncooked'),
('9a6490b0-e2e5-5aea-bc00-46d3a216e5c5'::uuid, 'Quinoa', '1 cup'),
('9a6490b0-e2e5-5aea-bc00-46d3a216e5c5'::uuid, 'Onion', '1'),
('9a6490b0-e2e5-5aea-bc00-46d3a216e5c5'::uuid, 'Canned tomatoes', '2 (14.5 oz) cans'),
('95714ebb-ae76-5fb3-a3da-882595b267fa'::uuid, 'Black pepper', 'to taste'),
('95714ebb-ae76-5fb3-a3da-882595b267fa'::uuid, 'Lime', '1'),
('95714ebb-ae76-5fb3-a3da-882595b267fa'::uuid, 'Salt', 'to taste'),
('95714ebb-ae76-5fb3-a3da-882595b267fa'::uuid, 'Canned tomatoes', '2 (14.5 oz) cans'),
('95714ebb-ae76-5fb3-a3da-882595b267fa'::uuid, 'Bread', '4 slices'),
('95714ebb-ae76-5fb3-a3da-882595b267fa'::uuid, 'Quinoa', '1 cup'),
('95714ebb-ae76-5fb3-a3da-882595b267fa'::uuid, 'Olive oil', '2 tbsp'),
('95714ebb-ae76-5fb3-a3da-882595b267fa'::uuid, 'Cilantro', '1/4 cup'),
('95714ebb-ae76-5fb3-a3da-882595b267fa'::uuid, 'Onion', '1'),
('95714ebb-ae76-5fb3-a3da-882595b267fa'::uuid, 'Cheddar cheese', '1 cup'),
('95714ebb-ae76-5fb3-a3da-882595b267fa'::uuid, 'Garlic', '2 cloves'),
('501955a6-71c6-5fb7-83b3-d68f9b714477'::uuid, 'Cilantro', '1/4 cup'),
('501955a6-71c6-5fb7-83b3-d68f9b714477'::uuid, 'Onion', '1'),
('501955a6-71c6-5fb7-83b3-d68f9b714477'::uuid, 'Vegetable broth', '3 cups'),
('501955a6-71c6-5fb7-83b3-d68f9b714477'::uuid, 'Quinoa', '1 cup'),
('501955a6-71c6-5fb7-83b3-d68f9b714477'::uuid, 'Canned tomatoes', '2 (14.5 oz) cans'),
('501955a6-71c6-5fb7-83b3-d68f9b714477'::uuid, 'Salt', 'to taste'),
('501955a6-71c6-5fb7-83b3-d68f9b714477'::uuid, 'Cheddar cheese', '1 cup'),
('501955a6-71c6-5fb7-83b3-d68f9b714477'::uuid, 'Bread', '4 slices'),
('501955a6-71c6-5fb7-83b3-d68f9b714477'::uuid, 'Garlic', '2 cloves'),
('43b4da68-85b9-5379-831c-bd5d8dd4775b'::uuid, 'Black pepper', 'to taste'),
('43b4da68-85b9-5379-831c-bd5d8dd4775b'::uuid, 'Olive oil', '2 tbsp'),
('43b4da68-85b9-5379-831c-bd5d8dd4775b'::uuid, 'Garlic', '2 cloves'),
('43b4da68-85b9-5379-831c-bd5d8dd4775b'::uuid, 'Rice', '1 cup uncooked'),
('43b4da68-85b9-5379-831c-bd5d8dd4775b'::uuid, 'Lime', '1'),
('43b4da68-85b9-5379-831c-bd5d8dd4775b'::uuid, 'Salt', 'to taste'),
('43b4da68-85b9-5379-831c-bd5d8dd4775b'::uuid, 'Bread', '4 slices'),
('43b4da68-85b9-5379-831c-bd5d8dd4775b'::uuid, 'Vegetable broth', '3 cups'),
('43b4da68-85b9-5379-831c-bd5d8dd4775b'::uuid, 'Quinoa', '1 cup'),
('0ab19725-1a7c-5d79-9c96-799e59098665'::uuid, 'Cheddar cheese', '1 cup'),
('0ab19725-1a7c-5d79-9c96-799e59098665'::uuid, 'Cilantro', '1/4 cup'),
('0ab19725-1a7c-5d79-9c96-799e59098665'::uuid, 'Lime', '1'),
('0ab19725-1a7c-5d79-9c96-799e59098665'::uuid, 'Olive oil', '2 tbsp'),
('0ab19725-1a7c-5d79-9c96-799e59098665'::uuid, 'Black pepper', 'to taste'),
('0ab19725-1a7c-5d79-9c96-799e59098665'::uuid, 'Bread', '4 slices'),
('0ab19725-1a7c-5d79-9c96-799e59098665'::uuid, 'Salt', 'to taste'),
('0ab19725-1a7c-5d79-9c96-799e59098665'::uuid, 'Onion', '1'),
('0ab19725-1a7c-5d79-9c96-799e59098665'::uuid, 'Canned tomatoes', '2 (14.5 oz) cans'),
('0ab19725-1a7c-5d79-9c96-799e59098665'::uuid, 'Rice', '1 cup uncooked'),
('0ab19725-1a7c-5d79-9c96-799e59098665'::uuid, 'Quinoa', '1 cup'),
('0ae2d4b1-423d-5415-bea9-2b9d08fb5a22'::uuid, 'Olive oil', '2 tbsp'),
('0ae2d4b1-423d-5415-bea9-2b9d08fb5a22'::uuid, 'Garlic', '2 cloves'),
('0ae2d4b1-423d-5415-bea9-2b9d08fb5a22'::uuid, 'Quinoa', '1 cup'),
('0ae2d4b1-423d-5415-bea9-2b9d08fb5a22'::uuid, 'Bread', '4 slices'),
('0ae2d4b1-423d-5415-bea9-2b9d08fb5a22'::uuid, 'Rice', '1 cup uncooked'),
('0ae2d4b1-423d-5415-bea9-2b9d08fb5a22'::uuid, 'Black pepper', 'to taste'),
('0ae2d4b1-423d-5415-bea9-2b9d08fb5a22'::uuid, 'Cheddar cheese', '1 cup'),
('0ae2d4b1-423d-5415-bea9-2b9d08fb5a22'::uuid, 'Cilantro', '1/4 cup'),
('0ae2d4b1-423d-5415-bea9-2b9d08fb5a22'::uuid, 'Lime', '1'),
('0ae2d4b1-423d-5415-bea9-2b9d08fb5a22'::uuid, 'Onion', '1'),
('0ae2d4b1-423d-5415-bea9-2b9d08fb5a22'::uuid, 'Salt', 'to taste'),
('39df1c43-555e-53e5-9414-b2bb3a88237b'::uuid, 'Quinoa', '1 cup'),
('39df1c43-555e-53e5-9414-b2bb3a88237b'::uuid, 'Lime', '1'),
('39df1c43-555e-53e5-9414-b2bb3a88237b'::uuid, 'Canned tomatoes', '2 (14.5 oz) cans'),
('39df1c43-555e-53e5-9414-b2bb3a88237b'::uuid, 'Bread', '4 slices'),
('39df1c43-555e-53e5-9414-b2bb3a88237b'::uuid, 'Black pepper', 'to taste'),
('39df1c43-555e-53e5-9414-b2bb3a88237b'::uuid, 'Cheddar cheese', '1 cup'),
('39df1c43-555e-53e5-9414-b2bb3a88237b'::uuid, 'Salt', 'to taste'),
('39df1c43-555e-53e5-9414-b2bb3a88237b'::uuid, 'Vegetable broth', '3 cups'),
('39df1c43-555e-53e5-9414-b2bb3a88237b'::uuid, 'Olive oil', '2 tbsp'),
('39df1c43-555e-53e5-9414-b2bb3a88237b'::uuid, 'Onion', '1'),
('39df1c43-555e-53e5-9414-b2bb3a88237b'::uuid, 'Garlic', '2 cloves'),
('39df1c43-555e-53e5-9414-b2bb3a88237b'::uuid, 'Cilantro', '1/4 cup'),
('3b23f590-17da-51e6-8aff-c378e559841c'::uuid, 'Vegetable broth', '3 cups'),
('3b23f590-17da-51e6-8aff-c378e559841c'::uuid, 'Olive oil', '2 tbsp'),
('3b23f590-17da-51e6-8aff-c378e559841c'::uuid, 'Cilantro', '1/4 cup'),
('3b23f590-17da-51e6-8aff-c378e559841c'::uuid, 'Black pepper', 'to taste'),
('3b23f590-17da-51e6-8aff-c378e559841c'::uuid, 'Quinoa', '1 cup'),
('3b23f590-17da-51e6-8aff-c378e559841c'::uuid, 'Bread', '4 slices'),
('3b23f590-17da-51e6-8aff-c378e559841c'::uuid, 'Canned tomatoes', '2 (14.5 oz) cans'),
('3b23f590-17da-51e6-8aff-c378e559841c'::uuid, 'Onion', '1'),
('3b23f590-17da-51e6-8aff-c378e559841c'::uuid, 'Garlic', '2 cloves'),
('3b23f590-17da-51e6-8aff-c378e559841c'::uuid, 'Salt', 'to taste'),
('3b23f590-17da-51e6-8aff-c378e559841c'::uuid, 'Rice', '1 cup uncooked'),
('a176f573-5785-5cd4-bd4a-c7ad0d0b2469'::uuid, 'Olive oil', '2 tbsp'),
('a176f573-5785-5cd4-bd4a-c7ad0d0b2469'::uuid, 'Vegetable broth', '3 cups'),
('a176f573-5785-5cd4-bd4a-c7ad0d0b2469'::uuid, 'Canned tomatoes', '2 (14.5 oz) cans'),
('a176f573-5785-5cd4-bd4a-c7ad0d0b2469'::uuid, 'Cilantro', '1/4 cup'),
('a176f573-5785-5cd4-bd4a-c7ad0d0b2469'::uuid, 'Quinoa', '1 cup'),
('a176f573-5785-5cd4-bd4a-c7ad0d0b2469'::uuid, 'Onion', '1'),
('a176f573-5785-5cd4-bd4a-c7ad0d0b2469'::uuid, 'Bread', '4 slices'),
('a176f573-5785-5cd4-bd4a-c7ad0d0b2469'::uuid, 'Cheddar cheese', '1 cup'),
('a176f573-5785-5cd4-bd4a-c7ad0d0b2469'::uuid, 'Lime', '1'),
('a176f573-5785-5cd4-bd4a-c7ad0d0b2469'::uuid, 'Salt', 'to taste'),
('fdd74c62-5009-53d9-bc37-5203f4a6fe6d'::uuid, 'Butter', '2 tbsp'),
('fdd74c62-5009-53d9-bc37-5203f4a6fe6d'::uuid, 'Crushed tomatoes', '1 (28 oz) can'),
('fdd74c62-5009-53d9-bc37-5203f4a6fe6d'::uuid, 'Olive oil', '2 tbsp'),
('fdd74c62-5009-53d9-bc37-5203f4a6fe6d'::uuid, 'Mozzarella', '1 cup'),
('fdd74c62-5009-53d9-bc37-5203f4a6fe6d'::uuid, 'Heavy cream', '1/2 cup'),
('fdd74c62-5009-53d9-bc37-5203f4a6fe6d'::uuid, 'Parmesan', '1/2 cup'),
('fdd74c62-5009-53d9-bc37-5203f4a6fe6d'::uuid, 'Ricotta', '1/2 cup'),
('fdd74c62-5009-53d9-bc37-5203f4a6fe6d'::uuid, 'Salt', 'to taste'),
('fdd74c62-5009-53d9-bc37-5203f4a6fe6d'::uuid, 'Pasta', '10 oz'),
('7bc5b32b-6e41-562b-a1e0-57acb7e52248'::uuid, 'Black pepper', 'to taste'),
('7bc5b32b-6e41-562b-a1e0-57acb7e52248'::uuid, 'Mozzarella', '1 cup'),
('7bc5b32b-6e41-562b-a1e0-57acb7e52248'::uuid, 'Salt', 'to taste'),
('7bc5b32b-6e41-562b-a1e0-57acb7e52248'::uuid, 'Frozen peas', '1 cup'),
('7bc5b32b-6e41-562b-a1e0-57acb7e52248'::uuid, 'Pasta', '10 oz'),
('7bc5b32b-6e41-562b-a1e0-57acb7e52248'::uuid, 'Butter', '2 tbsp'),
('7bc5b32b-6e41-562b-a1e0-57acb7e52248'::uuid, 'Ricotta', '1/2 cup'),
('7bc5b32b-6e41-562b-a1e0-57acb7e52248'::uuid, 'Heavy cream', '1/2 cup'),
('7bc5b32b-6e41-562b-a1e0-57acb7e52248'::uuid, 'Garlic', '3 cloves'),
('23f9d62b-3156-58eb-8e9b-c570307482cb'::uuid, 'Salt', 'to taste'),
('23f9d62b-3156-58eb-8e9b-c570307482cb'::uuid, 'Mozzarella', '1 cup'),
('23f9d62b-3156-58eb-8e9b-c570307482cb'::uuid, 'Black pepper', 'to taste'),
('23f9d62b-3156-58eb-8e9b-c570307482cb'::uuid, 'Pesto', '1/3 cup'),
('23f9d62b-3156-58eb-8e9b-c570307482cb'::uuid, 'Heavy cream', '1/2 cup'),
('23f9d62b-3156-58eb-8e9b-c570307482cb'::uuid, 'Crushed tomatoes', '1 (28 oz) can'),
('23f9d62b-3156-58eb-8e9b-c570307482cb'::uuid, 'Ricotta', '1/2 cup'),
('23f9d62b-3156-58eb-8e9b-c570307482cb'::uuid, 'Olive oil', '2 tbsp'),
('23f9d62b-3156-58eb-8e9b-c570307482cb'::uuid, 'Garlic', '3 cloves'),
('23f9d62b-3156-58eb-8e9b-c570307482cb'::uuid, 'Frozen peas', '1 cup'),
('6f4ee5d9-05df-5a2b-955b-2e75877e2539'::uuid, 'Pasta', '10 oz'),
('6f4ee5d9-05df-5a2b-955b-2e75877e2539'::uuid, 'Mozzarella', '1 cup'),
('6f4ee5d9-05df-5a2b-955b-2e75877e2539'::uuid, 'Parmesan', '1/2 cup'),
('6f4ee5d9-05df-5a2b-955b-2e75877e2539'::uuid, 'Heavy cream', '1/2 cup'),
('6f4ee5d9-05df-5a2b-955b-2e75877e2539'::uuid, 'Salt', 'to taste'),
('6f4ee5d9-05df-5a2b-955b-2e75877e2539'::uuid, 'Garlic', '3 cloves'),
('6f4ee5d9-05df-5a2b-955b-2e75877e2539'::uuid, 'Ricotta', '1/2 cup'),
('6f4ee5d9-05df-5a2b-955b-2e75877e2539'::uuid, 'Olive oil', '2 tbsp'),
('6f4ee5d9-05df-5a2b-955b-2e75877e2539'::uuid, 'Basil', '1/2 cup'),
('6f4ee5d9-05df-5a2b-955b-2e75877e2539'::uuid, 'Butter', '2 tbsp'),
('6f4ee5d9-05df-5a2b-955b-2e75877e2539'::uuid, 'Black pepper', 'to taste'),
('9bc0b6ff-d7ef-5151-b1a1-13c27dc3b4ac'::uuid, 'Frozen peas', '1 cup'),
('9bc0b6ff-d7ef-5151-b1a1-13c27dc3b4ac'::uuid, 'Black pepper', 'to taste'),
('9bc0b6ff-d7ef-5151-b1a1-13c27dc3b4ac'::uuid, 'Heavy cream', '1/2 cup'),
('9bc0b6ff-d7ef-5151-b1a1-13c27dc3b4ac'::uuid, 'Pasta', '10 oz'),
('9bc0b6ff-d7ef-5151-b1a1-13c27dc3b4ac'::uuid, 'Olive oil', '2 tbsp'),
('9bc0b6ff-d7ef-5151-b1a1-13c27dc3b4ac'::uuid, 'Mozzarella', '1 cup'),
('9bc0b6ff-d7ef-5151-b1a1-13c27dc3b4ac'::uuid, 'Garlic', '3 cloves'),
('9bc0b6ff-d7ef-5151-b1a1-13c27dc3b4ac'::uuid, 'Pesto', '1/3 cup'),
('9bc0b6ff-d7ef-5151-b1a1-13c27dc3b4ac'::uuid, 'Crushed tomatoes', '1 (28 oz) can'),
('9bc0b6ff-d7ef-5151-b1a1-13c27dc3b4ac'::uuid, 'Salt', 'to taste'),
('9bc0b6ff-d7ef-5151-b1a1-13c27dc3b4ac'::uuid, 'Butter', '2 tbsp'),
('12e9d911-c173-5dab-8b0b-12f189a2b5de'::uuid, 'Mozzarella', '1 cup'),
('12e9d911-c173-5dab-8b0b-12f189a2b5de'::uuid, 'Parmesan', '1/2 cup'),
('12e9d911-c173-5dab-8b0b-12f189a2b5de'::uuid, 'Crushed tomatoes', '1 (28 oz) can'),
('12e9d911-c173-5dab-8b0b-12f189a2b5de'::uuid, 'Ricotta', '1/2 cup'),
('12e9d911-c173-5dab-8b0b-12f189a2b5de'::uuid, 'Salt', 'to taste'),
('12e9d911-c173-5dab-8b0b-12f189a2b5de'::uuid, 'Frozen peas', '1 cup'),
('12e9d911-c173-5dab-8b0b-12f189a2b5de'::uuid, 'Black pepper', 'to taste'),
('12e9d911-c173-5dab-8b0b-12f189a2b5de'::uuid, 'Garlic', '3 cloves'),
('12e9d911-c173-5dab-8b0b-12f189a2b5de'::uuid, 'Butter', '2 tbsp'),
('12e9d911-c173-5dab-8b0b-12f189a2b5de'::uuid, 'Heavy cream', '1/2 cup'),
('f7b2162a-c42d-50c5-b2d0-918661f17e44'::uuid, 'Ricotta', '1/2 cup'),
('f7b2162a-c42d-50c5-b2d0-918661f17e44'::uuid, 'Pasta', '10 oz'),
('f7b2162a-c42d-50c5-b2d0-918661f17e44'::uuid, 'Butter', '2 tbsp'),
('f7b2162a-c42d-50c5-b2d0-918661f17e44'::uuid, 'Pesto', '1/3 cup'),
('f7b2162a-c42d-50c5-b2d0-918661f17e44'::uuid, 'Salt', 'to taste'),
('f7b2162a-c42d-50c5-b2d0-918661f17e44'::uuid, 'Parmesan', '1/2 cup'),
('f7b2162a-c42d-50c5-b2d0-918661f17e44'::uuid, 'Crushed tomatoes', '1 (28 oz) can'),
('f7b2162a-c42d-50c5-b2d0-918661f17e44'::uuid, 'Heavy cream', '1/2 cup'),
('9c3c5294-948c-5aa4-a85e-2eddd9957919'::uuid, 'Salt', 'to taste'),
('9c3c5294-948c-5aa4-a85e-2eddd9957919'::uuid, 'Heavy cream', '1/2 cup'),
('9c3c5294-948c-5aa4-a85e-2eddd9957919'::uuid, 'Pasta', '10 oz'),
('9c3c5294-948c-5aa4-a85e-2eddd9957919'::uuid, 'Butter', '2 tbsp'),
('9c3c5294-948c-5aa4-a85e-2eddd9957919'::uuid, 'Crushed tomatoes', '1 (28 oz) can'),
('9c3c5294-948c-5aa4-a85e-2eddd9957919'::uuid, 'Frozen peas', '1 cup'),
('9c3c5294-948c-5aa4-a85e-2eddd9957919'::uuid, 'Black pepper', 'to taste'),
('9c3c5294-948c-5aa4-a85e-2eddd9957919'::uuid, 'Basil', '1/2 cup'),
('a8b913c4-a147-55fe-a34f-a440058cf6dd'::uuid, 'Crushed tomatoes', '1 (28 oz) can'),
('a8b913c4-a147-55fe-a34f-a440058cf6dd'::uuid, 'Garlic', '3 cloves'),
('a8b913c4-a147-55fe-a34f-a440058cf6dd'::uuid, 'Ricotta', '1/2 cup'),
('a8b913c4-a147-55fe-a34f-a440058cf6dd'::uuid, 'Heavy cream', '1/2 cup'),
('a8b913c4-a147-55fe-a34f-a440058cf6dd'::uuid, 'Olive oil', '2 tbsp'),
('a8b913c4-a147-55fe-a34f-a440058cf6dd'::uuid, 'Black pepper', 'to taste'),
('a8b913c4-a147-55fe-a34f-a440058cf6dd'::uuid, 'Butter', '2 tbsp'),
('a8b913c4-a147-55fe-a34f-a440058cf6dd'::uuid, 'Parmesan', '1/2 cup'),
('a8b913c4-a147-55fe-a34f-a440058cf6dd'::uuid, 'Salt', 'to taste'),
('7faaf8b8-ec40-51a5-a5f5-dfe7eed91a83'::uuid, 'Garlic', '3 cloves'),
('7faaf8b8-ec40-51a5-a5f5-dfe7eed91a83'::uuid, 'Frozen peas', '1 cup'),
('7faaf8b8-ec40-51a5-a5f5-dfe7eed91a83'::uuid, 'Salt', 'to taste'),
('7faaf8b8-ec40-51a5-a5f5-dfe7eed91a83'::uuid, 'Olive oil', '2 tbsp'),
('7faaf8b8-ec40-51a5-a5f5-dfe7eed91a83'::uuid, 'Black pepper', 'to taste'),
('7faaf8b8-ec40-51a5-a5f5-dfe7eed91a83'::uuid, 'Basil', '1/2 cup'),
('7faaf8b8-ec40-51a5-a5f5-dfe7eed91a83'::uuid, 'Heavy cream', '1/2 cup'),
('7faaf8b8-ec40-51a5-a5f5-dfe7eed91a83'::uuid, 'Pesto', '1/3 cup'),
('7faaf8b8-ec40-51a5-a5f5-dfe7eed91a83'::uuid, 'Pasta', '10 oz'),
('7faaf8b8-ec40-51a5-a5f5-dfe7eed91a83'::uuid, 'Butter', '2 tbsp'),
('2d2539b2-c2d6-50f0-977d-0e767a5f69ff'::uuid, 'Orange', '1'),
('2d2539b2-c2d6-50f0-977d-0e767a5f69ff'::uuid, 'Black pepper', 'to taste'),
('2d2539b2-c2d6-50f0-977d-0e767a5f69ff'::uuid, 'Lime', '1'),
('2d2539b2-c2d6-50f0-977d-0e767a5f69ff'::uuid, 'Garlic', '3 cloves'),
('2d2539b2-c2d6-50f0-977d-0e767a5f69ff'::uuid, 'Pork shoulder', '2 lb'),
('2d2539b2-c2d6-50f0-977d-0e767a5f69ff'::uuid, 'Carrots', '1/2 cup, diced'),
('2d2539b2-c2d6-50f0-977d-0e767a5f69ff'::uuid, 'Green onions', '1/4 cup'),
('2d2539b2-c2d6-50f0-977d-0e767a5f69ff'::uuid, 'Cumin', '2 tsp'),
('2d2539b2-c2d6-50f0-977d-0e767a5f69ff'::uuid, 'Salt', 'to taste'),
('2d2539b2-c2d6-50f0-977d-0e767a5f69ff'::uuid, 'Soy sauce', '3 tbsp'),
('2d2539b2-c2d6-50f0-977d-0e767a5f69ff'::uuid, 'Rice', '1 cup uncooked'),
('2d2539b2-c2d6-50f0-977d-0e767a5f69ff'::uuid, 'Honey', '2 tbsp'),
('19f403ca-20a3-50b9-abea-e64e7bb6017a'::uuid, 'Garlic', '3 cloves'),
('19f403ca-20a3-50b9-abea-e64e7bb6017a'::uuid, 'Pork shoulder', '2 lb'),
('19f403ca-20a3-50b9-abea-e64e7bb6017a'::uuid, 'Honey', '2 tbsp'),
('19f403ca-20a3-50b9-abea-e64e7bb6017a'::uuid, 'Frozen peas', '1/2 cup'),
('19f403ca-20a3-50b9-abea-e64e7bb6017a'::uuid, 'Rice', '1 cup uncooked'),
('19f403ca-20a3-50b9-abea-e64e7bb6017a'::uuid, 'Green onions', '1/4 cup'),
('19f403ca-20a3-50b9-abea-e64e7bb6017a'::uuid, 'Oregano', '1 tsp'),
('19f403ca-20a3-50b9-abea-e64e7bb6017a'::uuid, 'Salt', 'to taste'),
('19f403ca-20a3-50b9-abea-e64e7bb6017a'::uuid, 'Cumin', '2 tsp'),
('19f403ca-20a3-50b9-abea-e64e7bb6017a'::uuid, 'Lime', '1'),
('19f403ca-20a3-50b9-abea-e64e7bb6017a'::uuid, 'Pork chops', '4'),
('19f403ca-20a3-50b9-abea-e64e7bb6017a'::uuid, 'Black pepper', 'to taste'),
('8c61a08b-f5bd-5d7d-97d2-bf45f38d82e3'::uuid, 'Garlic', '3 cloves'),
('8c61a08b-f5bd-5d7d-97d2-bf45f38d82e3'::uuid, 'Cumin', '2 tsp'),
('8c61a08b-f5bd-5d7d-97d2-bf45f38d82e3'::uuid, 'Honey', '2 tbsp'),
('8c61a08b-f5bd-5d7d-97d2-bf45f38d82e3'::uuid, 'Green onions', '1/4 cup'),
('8c61a08b-f5bd-5d7d-97d2-bf45f38d82e3'::uuid, 'Eggs', '2'),
('8c61a08b-f5bd-5d7d-97d2-bf45f38d82e3'::uuid, 'Lime', '1'),
('8c61a08b-f5bd-5d7d-97d2-bf45f38d82e3'::uuid, 'Olive oil', '2 tbsp'),
('8c61a08b-f5bd-5d7d-97d2-bf45f38d82e3'::uuid, 'Rice', '1 cup uncooked'),
('8c61a08b-f5bd-5d7d-97d2-bf45f38d82e3'::uuid, 'Salt', 'to taste'),
('8c61a08b-f5bd-5d7d-97d2-bf45f38d82e3'::uuid, 'Soy sauce', '3 tbsp'),
('8c61a08b-f5bd-5d7d-97d2-bf45f38d82e3'::uuid, 'Orange', '1'),
('1dbd5559-297c-5246-bdcd-3bcb30f28a1a'::uuid, 'Salt', 'to taste'),
('1dbd5559-297c-5246-bdcd-3bcb30f28a1a'::uuid, 'Oregano', '1 tsp'),
('1dbd5559-297c-5246-bdcd-3bcb30f28a1a'::uuid, 'Rice', '1 cup uncooked'),
('1dbd5559-297c-5246-bdcd-3bcb30f28a1a'::uuid, 'Cumin', '2 tsp'),
('1dbd5559-297c-5246-bdcd-3bcb30f28a1a'::uuid, 'Lime', '1'),
('1dbd5559-297c-5246-bdcd-3bcb30f28a1a'::uuid, 'Carrots', '1/2 cup, diced'),
('1dbd5559-297c-5246-bdcd-3bcb30f28a1a'::uuid, 'Soy sauce', '3 tbsp'),
('1dbd5559-297c-5246-bdcd-3bcb30f28a1a'::uuid, 'Honey', '2 tbsp'),
('1dbd5559-297c-5246-bdcd-3bcb30f28a1a'::uuid, 'Garlic', '3 cloves'),
('1dbd5559-297c-5246-bdcd-3bcb30f28a1a'::uuid, 'Pork shoulder', '2 lb'),
('1dbd5559-297c-5246-bdcd-3bcb30f28a1a'::uuid, 'Olive oil', '2 tbsp'),
('1dbd5559-297c-5246-bdcd-3bcb30f28a1a'::uuid, 'Frozen peas', '1/2 cup'),
('02b74751-ac90-5faa-b6d1-52d88b482b0b'::uuid, 'Oregano', '1 tsp'),
('02b74751-ac90-5faa-b6d1-52d88b482b0b'::uuid, 'Orange', '1'),
('02b74751-ac90-5faa-b6d1-52d88b482b0b'::uuid, 'Honey', '2 tbsp'),
('02b74751-ac90-5faa-b6d1-52d88b482b0b'::uuid, 'Pork chops', '4'),
('02b74751-ac90-5faa-b6d1-52d88b482b0b'::uuid, 'Eggs', '2'),
('02b74751-ac90-5faa-b6d1-52d88b482b0b'::uuid, 'Black pepper', 'to taste'),
('02b74751-ac90-5faa-b6d1-52d88b482b0b'::uuid, 'Olive oil', '2 tbsp'),
('02b74751-ac90-5faa-b6d1-52d88b482b0b'::uuid, 'Salt', 'to taste'),
('02b74751-ac90-5faa-b6d1-52d88b482b0b'::uuid, 'Garlic', '3 cloves'),
('efe492bf-bbf3-521d-826d-44a457f39232'::uuid, 'Frozen peas', '1/2 cup'),
('efe492bf-bbf3-521d-826d-44a457f39232'::uuid, 'Green onions', '1/4 cup'),
('efe492bf-bbf3-521d-826d-44a457f39232'::uuid, 'Soy sauce', '3 tbsp'),
('efe492bf-bbf3-521d-826d-44a457f39232'::uuid, 'Carrots', '1/2 cup, diced'),
('efe492bf-bbf3-521d-826d-44a457f39232'::uuid, 'Lime', '1'),
('efe492bf-bbf3-521d-826d-44a457f39232'::uuid, 'Oregano', '1 tsp'),
('efe492bf-bbf3-521d-826d-44a457f39232'::uuid, 'Black pepper', 'to taste'),
('efe492bf-bbf3-521d-826d-44a457f39232'::uuid, 'Pork chops', '4'),
('efe492bf-bbf3-521d-826d-44a457f39232'::uuid, 'Salt', 'to taste'),
('efe492bf-bbf3-521d-826d-44a457f39232'::uuid, 'Garlic', '3 cloves'),
('4848736f-65b2-57ab-9f1b-62d7645832f9'::uuid, 'Eggs', '2'),
('4848736f-65b2-57ab-9f1b-62d7645832f9'::uuid, 'Cumin', '2 tsp'),
('4848736f-65b2-57ab-9f1b-62d7645832f9'::uuid, 'Black pepper', 'to taste'),
('4848736f-65b2-57ab-9f1b-62d7645832f9'::uuid, 'Frozen peas', '1/2 cup'),
('4848736f-65b2-57ab-9f1b-62d7645832f9'::uuid, 'Orange', '1'),
('4848736f-65b2-57ab-9f1b-62d7645832f9'::uuid, 'Garlic', '3 cloves'),
('4848736f-65b2-57ab-9f1b-62d7645832f9'::uuid, 'Oregano', '1 tsp'),
('4848736f-65b2-57ab-9f1b-62d7645832f9'::uuid, 'Salt', 'to taste'),
('c79c78a6-8b65-57c6-9ee7-7939ab956613'::uuid, 'Pork chops', '4'),
('c79c78a6-8b65-57c6-9ee7-7939ab956613'::uuid, 'Olive oil', '2 tbsp'),
('c79c78a6-8b65-57c6-9ee7-7939ab956613'::uuid, 'Soy sauce', '3 tbsp'),
('c79c78a6-8b65-57c6-9ee7-7939ab956613'::uuid, 'Garlic', '3 cloves'),
('c79c78a6-8b65-57c6-9ee7-7939ab956613'::uuid, 'Rice', '1 cup uncooked'),
('c79c78a6-8b65-57c6-9ee7-7939ab956613'::uuid, 'Lime', '1'),
('c79c78a6-8b65-57c6-9ee7-7939ab956613'::uuid, 'Cumin', '2 tsp'),
('c79c78a6-8b65-57c6-9ee7-7939ab956613'::uuid, 'Oregano', '1 tsp'),
('c79c78a6-8b65-57c6-9ee7-7939ab956613'::uuid, 'Honey', '2 tbsp'),
('c79c78a6-8b65-57c6-9ee7-7939ab956613'::uuid, 'Eggs', '2'),
('c79c78a6-8b65-57c6-9ee7-7939ab956613'::uuid, 'Salt', 'to taste'),
('613a8382-1091-565c-86cc-37022bca1946'::uuid, 'Lime', '1'),
('613a8382-1091-565c-86cc-37022bca1946'::uuid, 'Cumin', '2 tsp'),
('613a8382-1091-565c-86cc-37022bca1946'::uuid, 'Black pepper', 'to taste'),
('613a8382-1091-565c-86cc-37022bca1946'::uuid, 'Frozen peas', '1/2 cup'),
('613a8382-1091-565c-86cc-37022bca1946'::uuid, 'Honey', '2 tbsp'),
('613a8382-1091-565c-86cc-37022bca1946'::uuid, 'Orange', '1'),
('613a8382-1091-565c-86cc-37022bca1946'::uuid, 'Eggs', '2'),
('613a8382-1091-565c-86cc-37022bca1946'::uuid, 'Soy sauce', '3 tbsp'),
('613a8382-1091-565c-86cc-37022bca1946'::uuid, 'Carrots', '1/2 cup, diced'),
('613a8382-1091-565c-86cc-37022bca1946'::uuid, 'Salt', 'to taste'),
('613a8382-1091-565c-86cc-37022bca1946'::uuid, 'Pork shoulder', '2 lb'),
('613a8382-1091-565c-86cc-37022bca1946'::uuid, 'Pork chops', '4'),
('110894b7-adea-507a-bf70-80cea261d9b6'::uuid, 'Green onions', '1/4 cup'),
('110894b7-adea-507a-bf70-80cea261d9b6'::uuid, 'Garlic', '3 cloves'),
('110894b7-adea-507a-bf70-80cea261d9b6'::uuid, 'Lime', '1'),
('110894b7-adea-507a-bf70-80cea261d9b6'::uuid, 'Frozen peas', '1/2 cup'),
('110894b7-adea-507a-bf70-80cea261d9b6'::uuid, 'Orange', '1'),
('110894b7-adea-507a-bf70-80cea261d9b6'::uuid, 'Oregano', '1 tsp'),
('110894b7-adea-507a-bf70-80cea261d9b6'::uuid, 'Olive oil', '2 tbsp'),
('110894b7-adea-507a-bf70-80cea261d9b6'::uuid, 'Salt', 'to taste'),
('110894b7-adea-507a-bf70-80cea261d9b6'::uuid, 'Carrots', '1/2 cup, diced'),
('04f148d1-ab13-5dbe-93c9-a3acce81d221'::uuid, 'Lemon', '1'),
('04f148d1-ab13-5dbe-93c9-a3acce81d221'::uuid, 'Soy sauce', '2 tbsp'),
('04f148d1-ab13-5dbe-93c9-a3acce81d221'::uuid, 'Sour cream', '1/3 cup'),
('04f148d1-ab13-5dbe-93c9-a3acce81d221'::uuid, 'Tortillas', '8'),
('04f148d1-ab13-5dbe-93c9-a3acce81d221'::uuid, 'Sesame oil', '1 tsp'),
('04f148d1-ab13-5dbe-93c9-a3acce81d221'::uuid, 'Black pepper', 'to taste'),
('04f148d1-ab13-5dbe-93c9-a3acce81d221'::uuid, 'Shrimp', '1 lb'),
('04f148d1-ab13-5dbe-93c9-a3acce81d221'::uuid, 'Olive oil', '2 tbsp'),
('04f148d1-ab13-5dbe-93c9-a3acce81d221'::uuid, 'Canned tuna', '1 (5 oz) can'),
('04f148d1-ab13-5dbe-93c9-a3acce81d221'::uuid, 'Salt', 'to taste'),
('04f148d1-ab13-5dbe-93c9-a3acce81d221'::uuid, 'Asparagus', '1 bunch'),
('32cc11a8-c195-513c-8fa4-44e846b56933'::uuid, 'Lemon', '1'),
('32cc11a8-c195-513c-8fa4-44e846b56933'::uuid, 'Salmon fillets', '2 (6 oz)'),
('32cc11a8-c195-513c-8fa4-44e846b56933'::uuid, 'Soy sauce', '2 tbsp'),
('32cc11a8-c195-513c-8fa4-44e846b56933'::uuid, 'Tortillas', '8'),
('32cc11a8-c195-513c-8fa4-44e846b56933'::uuid, 'Canned tuna', '1 (5 oz) can'),
('32cc11a8-c195-513c-8fa4-44e846b56933'::uuid, 'Black pepper', 'to taste'),
('32cc11a8-c195-513c-8fa4-44e846b56933'::uuid, 'Salt', 'to taste'),
('32cc11a8-c195-513c-8fa4-44e846b56933'::uuid, 'Sesame oil', '1 tsp'),
('32cc11a8-c195-513c-8fa4-44e846b56933'::uuid, 'Sour cream', '1/3 cup'),
('f5f314c0-a8d4-50e8-a2d8-997bf26840f8'::uuid, 'Soy sauce', '2 tbsp'),
('f5f314c0-a8d4-50e8-a2d8-997bf26840f8'::uuid, 'Canned tuna', '1 (5 oz) can'),
('f5f314c0-a8d4-50e8-a2d8-997bf26840f8'::uuid, 'Rice', '1 cup uncooked'),
('f5f314c0-a8d4-50e8-a2d8-997bf26840f8'::uuid, 'Cabbage slaw mix', '2 cups'),
('f5f314c0-a8d4-50e8-a2d8-997bf26840f8'::uuid, 'Shrimp', '1 lb'),
('f5f314c0-a8d4-50e8-a2d8-997bf26840f8'::uuid, 'Salmon fillets', '2 (6 oz)'),
('f5f314c0-a8d4-50e8-a2d8-997bf26840f8'::uuid, 'Lemon', '1'),
('f5f314c0-a8d4-50e8-a2d8-997bf26840f8'::uuid, 'Black pepper', 'to taste'),
('f5f314c0-a8d4-50e8-a2d8-997bf26840f8'::uuid, 'Olive oil', '2 tbsp'),
('f5f314c0-a8d4-50e8-a2d8-997bf26840f8'::uuid, 'Salt', 'to taste'),
('c355d867-7a68-58d3-b389-e8d27704251d'::uuid, 'Canned tuna', '1 (5 oz) can'),
('c355d867-7a68-58d3-b389-e8d27704251d'::uuid, 'Salmon fillets', '2 (6 oz)'),
('c355d867-7a68-58d3-b389-e8d27704251d'::uuid, 'Lemon', '1'),
('c355d867-7a68-58d3-b389-e8d27704251d'::uuid, 'Cabbage slaw mix', '2 cups'),
('c355d867-7a68-58d3-b389-e8d27704251d'::uuid, 'Shrimp', '1 lb'),
('c355d867-7a68-58d3-b389-e8d27704251d'::uuid, 'Rice', '1 cup uncooked'),
('c355d867-7a68-58d3-b389-e8d27704251d'::uuid, 'Asparagus', '1 bunch'),
('c355d867-7a68-58d3-b389-e8d27704251d'::uuid, 'Soy sauce', '2 tbsp'),
('c355d867-7a68-58d3-b389-e8d27704251d'::uuid, 'Salt', 'to taste'),
('c355d867-7a68-58d3-b389-e8d27704251d'::uuid, 'Sesame oil', '1 tsp'),
('9501d650-e0c7-545c-8a26-bec5b953a829'::uuid, 'Canned tuna', '1 (5 oz) can'),
('9501d650-e0c7-545c-8a26-bec5b953a829'::uuid, 'Black pepper', 'to taste'),
('9501d650-e0c7-545c-8a26-bec5b953a829'::uuid, 'Sesame oil', '1 tsp'),
('9501d650-e0c7-545c-8a26-bec5b953a829'::uuid, 'Sour cream', '1/3 cup'),
('9501d650-e0c7-545c-8a26-bec5b953a829'::uuid, 'Lemon', '1'),
('9501d650-e0c7-545c-8a26-bec5b953a829'::uuid, 'Shrimp', '1 lb'),
('9501d650-e0c7-545c-8a26-bec5b953a829'::uuid, 'Olive oil', '2 tbsp'),
('9501d650-e0c7-545c-8a26-bec5b953a829'::uuid, 'Salt', 'to taste'),
('4a2111a3-f682-511b-b4c4-20b3c6c3eb1b'::uuid, 'Sesame oil', '1 tsp'),
('4a2111a3-f682-511b-b4c4-20b3c6c3eb1b'::uuid, 'Asparagus', '1 bunch'),
('4a2111a3-f682-511b-b4c4-20b3c6c3eb1b'::uuid, 'Sour cream', '1/3 cup'),
('4a2111a3-f682-511b-b4c4-20b3c6c3eb1b'::uuid, 'Black pepper', 'to taste'),
('4a2111a3-f682-511b-b4c4-20b3c6c3eb1b'::uuid, 'Soy sauce', '2 tbsp'),
('4a2111a3-f682-511b-b4c4-20b3c6c3eb1b'::uuid, 'Olive oil', '2 tbsp'),
('4a2111a3-f682-511b-b4c4-20b3c6c3eb1b'::uuid, 'Tortillas', '8'),
('4a2111a3-f682-511b-b4c4-20b3c6c3eb1b'::uuid, 'Salt', 'to taste'),
('4a2111a3-f682-511b-b4c4-20b3c6c3eb1b'::uuid, 'Rice', '1 cup uncooked'),
('4a2111a3-f682-511b-b4c4-20b3c6c3eb1b'::uuid, 'Cabbage slaw mix', '2 cups'),
('4a2111a3-f682-511b-b4c4-20b3c6c3eb1b'::uuid, 'Canned tuna', '1 (5 oz) can'),
('4a2111a3-f682-511b-b4c4-20b3c6c3eb1b'::uuid, 'Garlic', '2 cloves'),
('cf4d26e6-fd86-5209-91ec-e0ff3e7410bd'::uuid, 'Sour cream', '1/3 cup'),
('cf4d26e6-fd86-5209-91ec-e0ff3e7410bd'::uuid, 'Black pepper', 'to taste'),
('cf4d26e6-fd86-5209-91ec-e0ff3e7410bd'::uuid, 'Salmon fillets', '2 (6 oz)'),
('cf4d26e6-fd86-5209-91ec-e0ff3e7410bd'::uuid, 'Salt', 'to taste'),
('cf4d26e6-fd86-5209-91ec-e0ff3e7410bd'::uuid, 'Shrimp', '1 lb'),
('cf4d26e6-fd86-5209-91ec-e0ff3e7410bd'::uuid, 'Olive oil', '2 tbsp'),
('cf4d26e6-fd86-5209-91ec-e0ff3e7410bd'::uuid, 'Soy sauce', '2 tbsp'),
('cf4d26e6-fd86-5209-91ec-e0ff3e7410bd'::uuid, 'Lemon', '1'),
('cf4d26e6-fd86-5209-91ec-e0ff3e7410bd'::uuid, 'Garlic', '2 cloves'),
('cf4d26e6-fd86-5209-91ec-e0ff3e7410bd'::uuid, 'Sesame oil', '1 tsp'),
('e7e682cc-cd15-544c-8b21-b67cfd4d8aa0'::uuid, 'Olive oil', '2 tbsp'),
('e7e682cc-cd15-544c-8b21-b67cfd4d8aa0'::uuid, 'Canned tuna', '1 (5 oz) can'),
('e7e682cc-cd15-544c-8b21-b67cfd4d8aa0'::uuid, 'Tortillas', '8'),
('e7e682cc-cd15-544c-8b21-b67cfd4d8aa0'::uuid, 'Shrimp', '1 lb'),
('e7e682cc-cd15-544c-8b21-b67cfd4d8aa0'::uuid, 'Salt', 'to taste'),
('e7e682cc-cd15-544c-8b21-b67cfd4d8aa0'::uuid, 'Salmon fillets', '2 (6 oz)'),
('e7e682cc-cd15-544c-8b21-b67cfd4d8aa0'::uuid, 'Asparagus', '1 bunch'),
('e7e682cc-cd15-544c-8b21-b67cfd4d8aa0'::uuid, 'Cabbage slaw mix', '2 cups'),
('e7e682cc-cd15-544c-8b21-b67cfd4d8aa0'::uuid, 'Garlic', '2 cloves'),
('9fd3101f-ed19-5569-835d-f56e708fb0c4'::uuid, 'Cabbage slaw mix', '2 cups'),
('9fd3101f-ed19-5569-835d-f56e708fb0c4'::uuid, 'Lemon', '1'),
('9fd3101f-ed19-5569-835d-f56e708fb0c4'::uuid, 'Salmon fillets', '2 (6 oz)'),
('9fd3101f-ed19-5569-835d-f56e708fb0c4'::uuid, 'Olive oil', '2 tbsp'),
('9fd3101f-ed19-5569-835d-f56e708fb0c4'::uuid, 'Salt', 'to taste'),
('9fd3101f-ed19-5569-835d-f56e708fb0c4'::uuid, 'Rice', '1 cup uncooked'),
('9fd3101f-ed19-5569-835d-f56e708fb0c4'::uuid, 'Black pepper', 'to taste'),
('9fd3101f-ed19-5569-835d-f56e708fb0c4'::uuid, 'Shrimp', '1 lb'),
('9fd3101f-ed19-5569-835d-f56e708fb0c4'::uuid, 'Garlic', '2 cloves'),
('9fd3101f-ed19-5569-835d-f56e708fb0c4'::uuid, 'Tortillas', '8'),
('9fd3101f-ed19-5569-835d-f56e708fb0c4'::uuid, 'Canned tuna', '1 (5 oz) can'),
('9fd3101f-ed19-5569-835d-f56e708fb0c4'::uuid, 'Asparagus', '1 bunch'),
('a0f9ac24-7bd9-5b8c-a195-52f25a9176dc'::uuid, 'Shrimp', '1 lb'),
('a0f9ac24-7bd9-5b8c-a195-52f25a9176dc'::uuid, 'Tortillas', '8'),
('a0f9ac24-7bd9-5b8c-a195-52f25a9176dc'::uuid, 'Olive oil', '2 tbsp'),
('a0f9ac24-7bd9-5b8c-a195-52f25a9176dc'::uuid, 'Rice', '1 cup uncooked'),
('a0f9ac24-7bd9-5b8c-a195-52f25a9176dc'::uuid, 'Canned tuna', '1 (5 oz) can'),
('a0f9ac24-7bd9-5b8c-a195-52f25a9176dc'::uuid, 'Lemon', '1'),
('a0f9ac24-7bd9-5b8c-a195-52f25a9176dc'::uuid, 'Soy sauce', '2 tbsp'),
('a0f9ac24-7bd9-5b8c-a195-52f25a9176dc'::uuid, 'Black pepper', 'to taste'),
('a0f9ac24-7bd9-5b8c-a195-52f25a9176dc'::uuid, 'Salmon fillets', '2 (6 oz)'),
('a0f9ac24-7bd9-5b8c-a195-52f25a9176dc'::uuid, 'Cabbage slaw mix', '2 cups'),
('a0f9ac24-7bd9-5b8c-a195-52f25a9176dc'::uuid, 'Salt', 'to taste'),
('7478213c-93a6-52f6-ae7b-631ca7c1540a'::uuid, 'Black pepper', 'to taste'),
('7478213c-93a6-52f6-ae7b-631ca7c1540a'::uuid, 'Sugar', '1 tbsp'),
('7478213c-93a6-52f6-ae7b-631ca7c1540a'::uuid, 'Broccoli', '3 cups'),
('7478213c-93a6-52f6-ae7b-631ca7c1540a'::uuid, 'Mayonnaise', '1/2 cup'),
('7478213c-93a6-52f6-ae7b-631ca7c1540a'::uuid, 'Olive oil', '2 tbsp'),
('7478213c-93a6-52f6-ae7b-631ca7c1540a'::uuid, 'Salt', 'to taste'),
('7478213c-93a6-52f6-ae7b-631ca7c1540a'::uuid, 'Butter', '2 tbsp'),
('7478213c-93a6-52f6-ae7b-631ca7c1540a'::uuid, 'Carrots', '3'),
('7478213c-93a6-52f6-ae7b-631ca7c1540a'::uuid, 'Green beans', '1 lb'),
('7478213c-93a6-52f6-ae7b-631ca7c1540a'::uuid, 'Garlic', '3 cloves'),
('7478213c-93a6-52f6-ae7b-631ca7c1540a'::uuid, 'Cabbage', '4 cups, shredded'),
('7478213c-93a6-52f6-ae7b-631ca7c1540a'::uuid, 'Lemon', '1'),
('f9523376-1410-501a-844e-a8001137f566'::uuid, 'Sugar', '1 tbsp'),
('f9523376-1410-501a-844e-a8001137f566'::uuid, 'Baby potatoes', '1 1/2 lb'),
('f9523376-1410-501a-844e-a8001137f566'::uuid, 'Cabbage', '4 cups, shredded'),
('f9523376-1410-501a-844e-a8001137f566'::uuid, 'Carrots', '3'),
('f9523376-1410-501a-844e-a8001137f566'::uuid, 'Butter', '2 tbsp'),
('f9523376-1410-501a-844e-a8001137f566'::uuid, 'Broccoli', '3 cups'),
('f9523376-1410-501a-844e-a8001137f566'::uuid, 'Lemon', '1'),
('f9523376-1410-501a-844e-a8001137f566'::uuid, 'Salt', 'to taste'),
('50a24b51-1906-5cee-8df0-b75d5b5a46ac'::uuid, 'Salt', 'to taste'),
('50a24b51-1906-5cee-8df0-b75d5b5a46ac'::uuid, 'Sugar', '1 tbsp'),
('50a24b51-1906-5cee-8df0-b75d5b5a46ac'::uuid, 'Lemon', '1'),
('50a24b51-1906-5cee-8df0-b75d5b5a46ac'::uuid, 'Baby potatoes', '1 1/2 lb'),
('50a24b51-1906-5cee-8df0-b75d5b5a46ac'::uuid, 'Broccoli', '3 cups'),
('50a24b51-1906-5cee-8df0-b75d5b5a46ac'::uuid, 'Butter', '2 tbsp'),
('50a24b51-1906-5cee-8df0-b75d5b5a46ac'::uuid, 'Garlic', '3 cloves'),
('50a24b51-1906-5cee-8df0-b75d5b5a46ac'::uuid, 'Green beans', '1 lb'),
('50a24b51-1906-5cee-8df0-b75d5b5a46ac'::uuid, 'Black pepper', 'to taste'),
('efd3e5c0-5d3f-56e7-86af-d2e13b54b80f'::uuid, 'Broccoli', '3 cups'),
('efd3e5c0-5d3f-56e7-86af-d2e13b54b80f'::uuid, 'Garlic', '3 cloves'),
('efd3e5c0-5d3f-56e7-86af-d2e13b54b80f'::uuid, 'Sugar', '1 tbsp'),
('efd3e5c0-5d3f-56e7-86af-d2e13b54b80f'::uuid, 'Cabbage', '4 cups, shredded'),
('efd3e5c0-5d3f-56e7-86af-d2e13b54b80f'::uuid, 'Mayonnaise', '1/2 cup'),
('efd3e5c0-5d3f-56e7-86af-d2e13b54b80f'::uuid, 'Olive oil', '2 tbsp'),
('efd3e5c0-5d3f-56e7-86af-d2e13b54b80f'::uuid, 'Black pepper', 'to taste'),
('efd3e5c0-5d3f-56e7-86af-d2e13b54b80f'::uuid, 'Lemon', '1'),
('efd3e5c0-5d3f-56e7-86af-d2e13b54b80f'::uuid, 'Green beans', '1 lb'),
('efd3e5c0-5d3f-56e7-86af-d2e13b54b80f'::uuid, 'Salt', 'to taste'),
('efd3e5c0-5d3f-56e7-86af-d2e13b54b80f'::uuid, 'Butter', '2 tbsp'),
('98469b7e-a5dc-54f1-9147-0df4a34ba840'::uuid, 'Butter', '2 tbsp'),
('98469b7e-a5dc-54f1-9147-0df4a34ba840'::uuid, 'Garlic', '3 cloves'),
('98469b7e-a5dc-54f1-9147-0df4a34ba840'::uuid, 'Carrots', '3'),
('98469b7e-a5dc-54f1-9147-0df4a34ba840'::uuid, 'Cabbage', '4 cups, shredded'),
('98469b7e-a5dc-54f1-9147-0df4a34ba840'::uuid, 'Mayonnaise', '1/2 cup'),
('98469b7e-a5dc-54f1-9147-0df4a34ba840'::uuid, 'Broccoli', '3 cups'),
('98469b7e-a5dc-54f1-9147-0df4a34ba840'::uuid, 'Salt', 'to taste'),
('98469b7e-a5dc-54f1-9147-0df4a34ba840'::uuid, 'Sugar', '1 tbsp'),
('6da33140-98b8-5519-87cf-cffd8307c94c'::uuid, 'Baby potatoes', '1 1/2 lb'),
('6da33140-98b8-5519-87cf-cffd8307c94c'::uuid, 'Mayonnaise', '1/2 cup'),
('6da33140-98b8-5519-87cf-cffd8307c94c'::uuid, 'Olive oil', '2 tbsp'),
('6da33140-98b8-5519-87cf-cffd8307c94c'::uuid, 'Carrots', '3'),
('6da33140-98b8-5519-87cf-cffd8307c94c'::uuid, 'Broccoli', '3 cups'),
('6da33140-98b8-5519-87cf-cffd8307c94c'::uuid, 'Salt', 'to taste'),
('6da33140-98b8-5519-87cf-cffd8307c94c'::uuid, 'Black pepper', 'to taste'),
('6da33140-98b8-5519-87cf-cffd8307c94c'::uuid, 'Cabbage', '4 cups, shredded'),
('6da33140-98b8-5519-87cf-cffd8307c94c'::uuid, 'Green beans', '1 lb'),
('6da33140-98b8-5519-87cf-cffd8307c94c'::uuid, 'Sugar', '1 tbsp'),
('6da33140-98b8-5519-87cf-cffd8307c94c'::uuid, 'Butter', '2 tbsp'),
('6da33140-98b8-5519-87cf-cffd8307c94c'::uuid, 'Garlic', '3 cloves'),
('59265790-cb79-532d-b01c-fea147198fa0'::uuid, 'Mayonnaise', '1/2 cup'),
('59265790-cb79-532d-b01c-fea147198fa0'::uuid, 'Cabbage', '4 cups, shredded'),
('59265790-cb79-532d-b01c-fea147198fa0'::uuid, 'Baby potatoes', '1 1/2 lb'),
('59265790-cb79-532d-b01c-fea147198fa0'::uuid, 'Green beans', '1 lb'),
('59265790-cb79-532d-b01c-fea147198fa0'::uuid, 'Vinegar', '1 tbsp'),
('59265790-cb79-532d-b01c-fea147198fa0'::uuid, 'Carrots', '3'),
('59265790-cb79-532d-b01c-fea147198fa0'::uuid, 'Butter', '2 tbsp'),
('59265790-cb79-532d-b01c-fea147198fa0'::uuid, 'Sugar', '1 tbsp'),
('59265790-cb79-532d-b01c-fea147198fa0'::uuid, 'Salt', 'to taste'),
('59265790-cb79-532d-b01c-fea147198fa0'::uuid, 'Black pepper', 'to taste'),
('59265790-cb79-532d-b01c-fea147198fa0'::uuid, 'Olive oil', '2 tbsp'),
('59265790-cb79-532d-b01c-fea147198fa0'::uuid, 'Lemon', '1'),
('d7f36ed2-05a7-5063-8b50-d57157577820'::uuid, 'Garlic', '3 cloves'),
('d7f36ed2-05a7-5063-8b50-d57157577820'::uuid, 'Broccoli', '3 cups'),
('d7f36ed2-05a7-5063-8b50-d57157577820'::uuid, 'Mayonnaise', '1/2 cup'),
('d7f36ed2-05a7-5063-8b50-d57157577820'::uuid, 'Black pepper', 'to taste'),
('d7f36ed2-05a7-5063-8b50-d57157577820'::uuid, 'Olive oil', '2 tbsp'),
('d7f36ed2-05a7-5063-8b50-d57157577820'::uuid, 'Baby potatoes', '1 1/2 lb'),
('d7f36ed2-05a7-5063-8b50-d57157577820'::uuid, 'Green beans', '1 lb'),
('d7f36ed2-05a7-5063-8b50-d57157577820'::uuid, 'Carrots', '3'),
('d7f36ed2-05a7-5063-8b50-d57157577820'::uuid, 'Sugar', '1 tbsp'),
('d7f36ed2-05a7-5063-8b50-d57157577820'::uuid, 'Butter', '2 tbsp'),
('d7f36ed2-05a7-5063-8b50-d57157577820'::uuid, 'Vinegar', '1 tbsp'),
('d7f36ed2-05a7-5063-8b50-d57157577820'::uuid, 'Salt', 'to taste'),
('6c0cf6fb-1e64-5095-b4b6-ecf16b6a7ed0'::uuid, 'Olive oil', '2 tbsp'),
('6c0cf6fb-1e64-5095-b4b6-ecf16b6a7ed0'::uuid, 'Green beans', '1 lb'),
('6c0cf6fb-1e64-5095-b4b6-ecf16b6a7ed0'::uuid, 'Butter', '2 tbsp'),
('6c0cf6fb-1e64-5095-b4b6-ecf16b6a7ed0'::uuid, 'Mayonnaise', '1/2 cup'),
('6c0cf6fb-1e64-5095-b4b6-ecf16b6a7ed0'::uuid, 'Black pepper', 'to taste'),
('6c0cf6fb-1e64-5095-b4b6-ecf16b6a7ed0'::uuid, 'Garlic', '3 cloves'),
('6c0cf6fb-1e64-5095-b4b6-ecf16b6a7ed0'::uuid, 'Salt', 'to taste'),
('6c0cf6fb-1e64-5095-b4b6-ecf16b6a7ed0'::uuid, 'Carrots', '3'),
('6c0cf6fb-1e64-5095-b4b6-ecf16b6a7ed0'::uuid, 'Vinegar', '1 tbsp'),
('6c0cf6fb-1e64-5095-b4b6-ecf16b6a7ed0'::uuid, 'Broccoli', '3 cups'),
('6c0cf6fb-1e64-5095-b4b6-ecf16b6a7ed0'::uuid, 'Cabbage', '4 cups, shredded'),
('017ed580-e30a-5e7e-8fb4-33492b92ede8'::uuid, 'Butter', '2 tbsp'),
('017ed580-e30a-5e7e-8fb4-33492b92ede8'::uuid, 'Mayonnaise', '1/2 cup'),
('017ed580-e30a-5e7e-8fb4-33492b92ede8'::uuid, 'Lemon', '1'),
('017ed580-e30a-5e7e-8fb4-33492b92ede8'::uuid, 'Green beans', '1 lb'),
('017ed580-e30a-5e7e-8fb4-33492b92ede8'::uuid, 'Black pepper', 'to taste'),
('017ed580-e30a-5e7e-8fb4-33492b92ede8'::uuid, 'Vinegar', '1 tbsp'),
('017ed580-e30a-5e7e-8fb4-33492b92ede8'::uuid, 'Garlic', '3 cloves'),
('017ed580-e30a-5e7e-8fb4-33492b92ede8'::uuid, 'Salt', 'to taste'),
('017ed580-e30a-5e7e-8fb4-33492b92ede8'::uuid, 'Olive oil', '2 tbsp'),
('017ed580-e30a-5e7e-8fb4-33492b92ede8'::uuid, 'Broccoli', '3 cups'),
('017ed580-e30a-5e7e-8fb4-33492b92ede8'::uuid, 'Carrots', '3'),
('d8af8860-c8d5-5807-a91d-e5dc20b176c9'::uuid, 'Spinach', '3 cups'),
('d8af8860-c8d5-5807-a91d-e5dc20b176c9'::uuid, 'Rice vinegar', '1 tbsp'),
('d8af8860-c8d5-5807-a91d-e5dc20b176c9'::uuid, 'Salt', 'to taste'),
('d8af8860-c8d5-5807-a91d-e5dc20b176c9'::uuid, 'Coconut milk', '1 (13.5 oz) can'),
('d8af8860-c8d5-5807-a91d-e5dc20b176c9'::uuid, 'Olive oil', '2 tbsp'),
('d8af8860-c8d5-5807-a91d-e5dc20b176c9'::uuid, 'Black pepper', 'to taste'),
('d8af8860-c8d5-5807-a91d-e5dc20b176c9'::uuid, 'Garlic', '3 cloves'),
('d8af8860-c8d5-5807-a91d-e5dc20b176c9'::uuid, 'Chickpeas', '2 (15 oz) cans'),
('3e74bf9d-9695-5b01-8aac-7d98058185f2'::uuid, 'Rice', '1 cup uncooked'),
('3e74bf9d-9695-5b01-8aac-7d98058185f2'::uuid, 'Ginger', '1 tbsp'),
('3e74bf9d-9695-5b01-8aac-7d98058185f2'::uuid, 'Black pepper', 'to taste'),
('3e74bf9d-9695-5b01-8aac-7d98058185f2'::uuid, 'Salt', 'to taste'),
('3e74bf9d-9695-5b01-8aac-7d98058185f2'::uuid, 'Bell pepper', '2'),
('3e74bf9d-9695-5b01-8aac-7d98058185f2'::uuid, 'Diced tomatoes', '1 (14.5 oz) can'),
('3e74bf9d-9695-5b01-8aac-7d98058185f2'::uuid, 'Black beans', '1 (15 oz) can'),
('3e74bf9d-9695-5b01-8aac-7d98058185f2'::uuid, 'Soy sauce', '3 tbsp'),
('3e74bf9d-9695-5b01-8aac-7d98058185f2'::uuid, 'Spinach', '3 cups'),
('3e74bf9d-9695-5b01-8aac-7d98058185f2'::uuid, 'Rice vinegar', '1 tbsp'),
('c7684b92-957d-51d1-a0b1-043c9a933a3a'::uuid, 'Rice vinegar', '1 tbsp'),
('c7684b92-957d-51d1-a0b1-043c9a933a3a'::uuid, 'Diced tomatoes', '1 (14.5 oz) can'),
('c7684b92-957d-51d1-a0b1-043c9a933a3a'::uuid, 'Olive oil', '2 tbsp'),
('c7684b92-957d-51d1-a0b1-043c9a933a3a'::uuid, 'Onion', '1'),
('c7684b92-957d-51d1-a0b1-043c9a933a3a'::uuid, 'Soy sauce', '3 tbsp'),
('c7684b92-957d-51d1-a0b1-043c9a933a3a'::uuid, 'Black pepper', 'to taste'),
('c7684b92-957d-51d1-a0b1-043c9a933a3a'::uuid, 'Rice', '1 cup uncooked'),
('c7684b92-957d-51d1-a0b1-043c9a933a3a'::uuid, 'Coconut milk', '1 (13.5 oz) can'),
('c7684b92-957d-51d1-a0b1-043c9a933a3a'::uuid, 'Bell pepper', '2'),
('c7684b92-957d-51d1-a0b1-043c9a933a3a'::uuid, 'Salt', 'to taste'),
('e79d4e5e-b4e7-59a9-9954-ba1261ab9522'::uuid, 'Lentils', '1 cup'),
('e79d4e5e-b4e7-59a9-9954-ba1261ab9522'::uuid, 'Olive oil', '2 tbsp'),
('e79d4e5e-b4e7-59a9-9954-ba1261ab9522'::uuid, 'Soy sauce', '3 tbsp'),
('e79d4e5e-b4e7-59a9-9954-ba1261ab9522'::uuid, 'Diced tomatoes', '1 (14.5 oz) can'),
('e79d4e5e-b4e7-59a9-9954-ba1261ab9522'::uuid, 'Ginger', '1 tbsp'),
('e79d4e5e-b4e7-59a9-9954-ba1261ab9522'::uuid, 'Coconut milk', '1 (13.5 oz) can'),
('e79d4e5e-b4e7-59a9-9954-ba1261ab9522'::uuid, 'Bell pepper', '2'),
('e79d4e5e-b4e7-59a9-9954-ba1261ab9522'::uuid, 'Black beans', '1 (15 oz) can'),
('e79d4e5e-b4e7-59a9-9954-ba1261ab9522'::uuid, 'Rice vinegar', '1 tbsp'),
('e79d4e5e-b4e7-59a9-9954-ba1261ab9522'::uuid, 'Chickpeas', '2 (15 oz) cans'),
('e79d4e5e-b4e7-59a9-9954-ba1261ab9522'::uuid, 'Salt', 'to taste'),
('363c757b-b4d6-5f6d-896e-b532accfbc1e'::uuid, 'Bell pepper', '2'),
('363c757b-b4d6-5f6d-896e-b532accfbc1e'::uuid, 'Ginger', '1 tbsp'),
('363c757b-b4d6-5f6d-896e-b532accfbc1e'::uuid, 'Lentils', '1 cup'),
('363c757b-b4d6-5f6d-896e-b532accfbc1e'::uuid, 'Diced tomatoes', '1 (14.5 oz) can'),
('363c757b-b4d6-5f6d-896e-b532accfbc1e'::uuid, 'Salt', 'to taste'),
('363c757b-b4d6-5f6d-896e-b532accfbc1e'::uuid, 'Spinach', '3 cups'),
('363c757b-b4d6-5f6d-896e-b532accfbc1e'::uuid, 'Rice vinegar', '1 tbsp'),
('363c757b-b4d6-5f6d-896e-b532accfbc1e'::uuid, 'Chickpeas', '2 (15 oz) cans'),
('363c757b-b4d6-5f6d-896e-b532accfbc1e'::uuid, 'Black beans', '1 (15 oz) can'),
('363c757b-b4d6-5f6d-896e-b532accfbc1e'::uuid, 'Onion', '1'),
('363c757b-b4d6-5f6d-896e-b532accfbc1e'::uuid, 'Olive oil', '2 tbsp'),
('363c757b-b4d6-5f6d-896e-b532accfbc1e'::uuid, 'Soy sauce', '3 tbsp'),
('b2b8ba13-ff1e-5c33-8fcf-55af44ac3c2f'::uuid, 'Black pepper', 'to taste'),
('b2b8ba13-ff1e-5c33-8fcf-55af44ac3c2f'::uuid, 'Rice vinegar', '1 tbsp'),
('b2b8ba13-ff1e-5c33-8fcf-55af44ac3c2f'::uuid, 'Garlic', '3 cloves'),
('b2b8ba13-ff1e-5c33-8fcf-55af44ac3c2f'::uuid, 'Olive oil', '2 tbsp'),
('b2b8ba13-ff1e-5c33-8fcf-55af44ac3c2f'::uuid, 'Rice', '1 cup uncooked'),
('b2b8ba13-ff1e-5c33-8fcf-55af44ac3c2f'::uuid, 'Coconut milk', '1 (13.5 oz) can'),
('b2b8ba13-ff1e-5c33-8fcf-55af44ac3c2f'::uuid, 'Lentils', '1 cup'),
('b2b8ba13-ff1e-5c33-8fcf-55af44ac3c2f'::uuid, 'Salt', 'to taste'),
('1330a946-51ef-5fb9-8606-f197577377b4'::uuid, 'Rice vinegar', '1 tbsp'),
('1330a946-51ef-5fb9-8606-f197577377b4'::uuid, 'Olive oil', '2 tbsp'),
('1330a946-51ef-5fb9-8606-f197577377b4'::uuid, 'Bell pepper', '2'),
('1330a946-51ef-5fb9-8606-f197577377b4'::uuid, 'Soy sauce', '3 tbsp'),
('1330a946-51ef-5fb9-8606-f197577377b4'::uuid, 'Diced tomatoes', '1 (14.5 oz) can'),
('1330a946-51ef-5fb9-8606-f197577377b4'::uuid, 'Spinach', '3 cups'),
('1330a946-51ef-5fb9-8606-f197577377b4'::uuid, 'Chickpeas', '2 (15 oz) cans'),
('1330a946-51ef-5fb9-8606-f197577377b4'::uuid, 'Lentils', '1 cup'),
('1330a946-51ef-5fb9-8606-f197577377b4'::uuid, 'Onion', '1'),
('1330a946-51ef-5fb9-8606-f197577377b4'::uuid, 'Salt', 'to taste'),
('1330a946-51ef-5fb9-8606-f197577377b4'::uuid, 'Coconut milk', '1 (13.5 oz) can'),
('1330a946-51ef-5fb9-8606-f197577377b4'::uuid, 'Black beans', '1 (15 oz) can'),
('1a687bb8-fc57-5ac7-b7eb-03bd1800d54c'::uuid, 'Black beans', '1 (15 oz) can'),
('1a687bb8-fc57-5ac7-b7eb-03bd1800d54c'::uuid, 'Coconut milk', '1 (13.5 oz) can'),
('1a687bb8-fc57-5ac7-b7eb-03bd1800d54c'::uuid, 'Lentils', '1 cup'),
('1a687bb8-fc57-5ac7-b7eb-03bd1800d54c'::uuid, 'Rice vinegar', '1 tbsp'),
('1a687bb8-fc57-5ac7-b7eb-03bd1800d54c'::uuid, 'Bell pepper', '2'),
('1a687bb8-fc57-5ac7-b7eb-03bd1800d54c'::uuid, 'Ginger', '1 tbsp'),
('1a687bb8-fc57-5ac7-b7eb-03bd1800d54c'::uuid, 'Diced tomatoes', '1 (14.5 oz) can'),
('1a687bb8-fc57-5ac7-b7eb-03bd1800d54c'::uuid, 'Black pepper', 'to taste'),
('1a687bb8-fc57-5ac7-b7eb-03bd1800d54c'::uuid, 'Salt', 'to taste'),
('aa29d27e-6068-5ea5-b2ba-5a52ee67a09a'::uuid, 'Lentils', '1 cup'),
('aa29d27e-6068-5ea5-b2ba-5a52ee67a09a'::uuid, 'Olive oil', '2 tbsp'),
('aa29d27e-6068-5ea5-b2ba-5a52ee67a09a'::uuid, 'Rice vinegar', '1 tbsp'),
('aa29d27e-6068-5ea5-b2ba-5a52ee67a09a'::uuid, 'Chickpeas', '2 (15 oz) cans'),
('aa29d27e-6068-5ea5-b2ba-5a52ee67a09a'::uuid, 'Garlic', '3 cloves'),
('aa29d27e-6068-5ea5-b2ba-5a52ee67a09a'::uuid, 'Salt', 'to taste'),
('aa29d27e-6068-5ea5-b2ba-5a52ee67a09a'::uuid, 'Bell pepper', '2'),
('aa29d27e-6068-5ea5-b2ba-5a52ee67a09a'::uuid, 'Coconut milk', '1 (13.5 oz) can'),
('ef3185ab-78e1-5c2e-8ed1-8c9df2016fc2'::uuid, 'Diced tomatoes', '1 (14.5 oz) can'),
('ef3185ab-78e1-5c2e-8ed1-8c9df2016fc2'::uuid, 'Spinach', '3 cups'),
('ef3185ab-78e1-5c2e-8ed1-8c9df2016fc2'::uuid, 'Coconut milk', '1 (13.5 oz) can'),
('ef3185ab-78e1-5c2e-8ed1-8c9df2016fc2'::uuid, 'Rice', '1 cup uncooked'),
('ef3185ab-78e1-5c2e-8ed1-8c9df2016fc2'::uuid, 'Bell pepper', '2'),
('ef3185ab-78e1-5c2e-8ed1-8c9df2016fc2'::uuid, 'Black beans', '1 (15 oz) can'),
('ef3185ab-78e1-5c2e-8ed1-8c9df2016fc2'::uuid, 'Soy sauce', '3 tbsp'),
('ef3185ab-78e1-5c2e-8ed1-8c9df2016fc2'::uuid, 'Rice vinegar', '1 tbsp'),
('ef3185ab-78e1-5c2e-8ed1-8c9df2016fc2'::uuid, 'Salt', 'to taste'),
('fec99184-5e4b-573d-9c9b-34edb5bb2732'::uuid, 'Baking powder', '2 tsp'),
('fec99184-5e4b-573d-9c9b-34edb5bb2732'::uuid, 'Honey', '1 tbsp'),
('fec99184-5e4b-573d-9c9b-34edb5bb2732'::uuid, 'Vanilla extract', '1 tsp'),
('fec99184-5e4b-573d-9c9b-34edb5bb2732'::uuid, 'Salt', 'pinch'),
('fec99184-5e4b-573d-9c9b-34edb5bb2732'::uuid, 'Avocado', '1'),
('fec99184-5e4b-573d-9c9b-34edb5bb2732'::uuid, 'Rolled oats', '1/2 cup'),
('fec99184-5e4b-573d-9c9b-34edb5bb2732'::uuid, 'Greek yogurt', '1 cup'),
('fec99184-5e4b-573d-9c9b-34edb5bb2732'::uuid, 'Cinnamon', '1/2 tsp'),
('fec99184-5e4b-573d-9c9b-34edb5bb2732'::uuid, 'Banana', '1'),
('fec99184-5e4b-573d-9c9b-34edb5bb2732'::uuid, 'Butter', '1 tbsp'),
('fec99184-5e4b-573d-9c9b-34edb5bb2732'::uuid, 'Sugar', '2 tbsp'),
('fec99184-5e4b-573d-9c9b-34edb5bb2732'::uuid, 'Berries', '1 cup'),
('93f134ef-df0b-5279-b22e-e279d5d737b6'::uuid, 'Salt', 'pinch'),
('93f134ef-df0b-5279-b22e-e279d5d737b6'::uuid, 'Avocado', '1'),
('93f134ef-df0b-5279-b22e-e279d5d737b6'::uuid, 'Berries', '1 cup'),
('93f134ef-df0b-5279-b22e-e279d5d737b6'::uuid, 'Milk', '1/2 cup'),
('93f134ef-df0b-5279-b22e-e279d5d737b6'::uuid, 'Honey', '1 tbsp'),
('93f134ef-df0b-5279-b22e-e279d5d737b6'::uuid, 'Vanilla extract', '1 tsp'),
('93f134ef-df0b-5279-b22e-e279d5d737b6'::uuid, 'Cinnamon', '1/2 tsp'),
('93f134ef-df0b-5279-b22e-e279d5d737b6'::uuid, 'Banana', '1'),
('540b4aad-2c28-5db5-ab30-b2de209d7cc0'::uuid, 'Bread', '2 slices'),
('540b4aad-2c28-5db5-ab30-b2de209d7cc0'::uuid, 'Salt', 'pinch'),
('540b4aad-2c28-5db5-ab30-b2de209d7cc0'::uuid, 'Avocado', '1'),
('540b4aad-2c28-5db5-ab30-b2de209d7cc0'::uuid, 'Greek yogurt', '1 cup'),
('540b4aad-2c28-5db5-ab30-b2de209d7cc0'::uuid, 'Chia seeds', '1 tbsp'),
('540b4aad-2c28-5db5-ab30-b2de209d7cc0'::uuid, 'Honey', '1 tbsp'),
('540b4aad-2c28-5db5-ab30-b2de209d7cc0'::uuid, 'Milk', '1/2 cup'),
('540b4aad-2c28-5db5-ab30-b2de209d7cc0'::uuid, 'Cinnamon', '1/2 tsp'),
('540b4aad-2c28-5db5-ab30-b2de209d7cc0'::uuid, 'Sugar', '2 tbsp'),
('540b4aad-2c28-5db5-ab30-b2de209d7cc0'::uuid, 'Berries', '1 cup'),
('540b4aad-2c28-5db5-ab30-b2de209d7cc0'::uuid, 'Banana', '1'),
('e152b51b-187a-51a7-b891-0c0f0b64fc90'::uuid, 'Eggs', '2'),
('e152b51b-187a-51a7-b891-0c0f0b64fc90'::uuid, 'Baking powder', '2 tsp'),
('e152b51b-187a-51a7-b891-0c0f0b64fc90'::uuid, 'Bread', '2 slices'),
('e152b51b-187a-51a7-b891-0c0f0b64fc90'::uuid, 'Granola', '1/2 cup'),
('e152b51b-187a-51a7-b891-0c0f0b64fc90'::uuid, 'Avocado', '1'),
('e152b51b-187a-51a7-b891-0c0f0b64fc90'::uuid, 'Butter', '1 tbsp'),
('e152b51b-187a-51a7-b891-0c0f0b64fc90'::uuid, 'Flour', '1 cup'),
('e152b51b-187a-51a7-b891-0c0f0b64fc90'::uuid, 'Salt', 'pinch'),
('e152b51b-187a-51a7-b891-0c0f0b64fc90'::uuid, 'Sugar', '2 tbsp'),
('e152b51b-187a-51a7-b891-0c0f0b64fc90'::uuid, 'Honey', '1 tbsp'),
('5546b970-f852-5b1b-96c7-bfdb825a8330'::uuid, 'Honey', '1 tbsp'),
('5546b970-f852-5b1b-96c7-bfdb825a8330'::uuid, 'Flour', '1 cup'),
('5546b970-f852-5b1b-96c7-bfdb825a8330'::uuid, 'Granola', '1/2 cup'),
('5546b970-f852-5b1b-96c7-bfdb825a8330'::uuid, 'Butter', '1 tbsp'),
('5546b970-f852-5b1b-96c7-bfdb825a8330'::uuid, 'Greek yogurt', '1 cup'),
('5546b970-f852-5b1b-96c7-bfdb825a8330'::uuid, 'Avocado', '1'),
('5546b970-f852-5b1b-96c7-bfdb825a8330'::uuid, 'Banana', '1'),
('5546b970-f852-5b1b-96c7-bfdb825a8330'::uuid, 'Salt', 'pinch'),
('5546b970-f852-5b1b-96c7-bfdb825a8330'::uuid, 'Rolled oats', '1/2 cup'),
('5546b970-f852-5b1b-96c7-bfdb825a8330'::uuid, 'Milk', '1/2 cup'),
('8926e00c-061b-5fe4-853c-79a529a2ffa8'::uuid, 'Avocado', '1'),
('8926e00c-061b-5fe4-853c-79a529a2ffa8'::uuid, 'Salt', 'pinch'),
('8926e00c-061b-5fe4-853c-79a529a2ffa8'::uuid, 'Milk', '1/2 cup'),
('8926e00c-061b-5fe4-853c-79a529a2ffa8'::uuid, 'Granola', '1/2 cup'),
('8926e00c-061b-5fe4-853c-79a529a2ffa8'::uuid, 'Berries', '1 cup'),
('8926e00c-061b-5fe4-853c-79a529a2ffa8'::uuid, 'Flour', '1 cup'),
('8926e00c-061b-5fe4-853c-79a529a2ffa8'::uuid, 'Rolled oats', '1/2 cup'),
('8926e00c-061b-5fe4-853c-79a529a2ffa8'::uuid, 'Vanilla extract', '1 tsp'),
('f6897b89-3037-5065-8ed0-fe235de96584'::uuid, 'Butter', '1 tbsp'),
('f6897b89-3037-5065-8ed0-fe235de96584'::uuid, 'Sugar', '2 tbsp'),
('f6897b89-3037-5065-8ed0-fe235de96584'::uuid, 'Berries', '1 cup'),
('f6897b89-3037-5065-8ed0-fe235de96584'::uuid, 'Greek yogurt', '1 cup'),
('f6897b89-3037-5065-8ed0-fe235de96584'::uuid, 'Rolled oats', '1/2 cup'),
('f6897b89-3037-5065-8ed0-fe235de96584'::uuid, 'Baking powder', '2 tsp'),
('f6897b89-3037-5065-8ed0-fe235de96584'::uuid, 'Bread', '2 slices'),
('f6897b89-3037-5065-8ed0-fe235de96584'::uuid, 'Vanilla extract', '1 tsp'),
('f6897b89-3037-5065-8ed0-fe235de96584'::uuid, 'Flour', '1 cup'),
('f6897b89-3037-5065-8ed0-fe235de96584'::uuid, 'Avocado', '1'),
('0a127797-69be-5039-938d-9b235208b2f3'::uuid, 'Berries', '1 cup'),
('0a127797-69be-5039-938d-9b235208b2f3'::uuid, 'Bread', '2 slices'),
('0a127797-69be-5039-938d-9b235208b2f3'::uuid, 'Vanilla extract', '1 tsp'),
('0a127797-69be-5039-938d-9b235208b2f3'::uuid, 'Baking powder', '2 tsp'),
('0a127797-69be-5039-938d-9b235208b2f3'::uuid, 'Banana', '1'),
('0a127797-69be-5039-938d-9b235208b2f3'::uuid, 'Flour', '1 cup'),
('0a127797-69be-5039-938d-9b235208b2f3'::uuid, 'Granola', '1/2 cup'),
('0a127797-69be-5039-938d-9b235208b2f3'::uuid, 'Milk', '1/2 cup'),
('0a127797-69be-5039-938d-9b235208b2f3'::uuid, 'Honey', '1 tbsp'),
('0a127797-69be-5039-938d-9b235208b2f3'::uuid, 'Butter', '1 tbsp'),
('c3ea2d31-7b34-5032-baed-fa71669cb614'::uuid, 'Honey', '1 tbsp'),
('c3ea2d31-7b34-5032-baed-fa71669cb614'::uuid, 'Baking powder', '2 tsp'),
('c3ea2d31-7b34-5032-baed-fa71669cb614'::uuid, 'Flour', '1 cup'),
('c3ea2d31-7b34-5032-baed-fa71669cb614'::uuid, 'Milk', '1/2 cup'),
('c3ea2d31-7b34-5032-baed-fa71669cb614'::uuid, 'Cinnamon', '1/2 tsp'),
('c3ea2d31-7b34-5032-baed-fa71669cb614'::uuid, 'Chia seeds', '1 tbsp'),
('c3ea2d31-7b34-5032-baed-fa71669cb614'::uuid, 'Eggs', '2'),
('c3ea2d31-7b34-5032-baed-fa71669cb614'::uuid, 'Banana', '1'),
('c3ea2d31-7b34-5032-baed-fa71669cb614'::uuid, 'Bread', '2 slices'),
('e5ff2a13-cc76-5dd2-bda1-53ed68ee5fd7'::uuid, 'Salt', 'pinch'),
('e5ff2a13-cc76-5dd2-bda1-53ed68ee5fd7'::uuid, 'Flour', '1 cup'),
('e5ff2a13-cc76-5dd2-bda1-53ed68ee5fd7'::uuid, 'Bread', '2 slices'),
('e5ff2a13-cc76-5dd2-bda1-53ed68ee5fd7'::uuid, 'Avocado', '1'),
('e5ff2a13-cc76-5dd2-bda1-53ed68ee5fd7'::uuid, 'Chia seeds', '1 tbsp'),
('e5ff2a13-cc76-5dd2-bda1-53ed68ee5fd7'::uuid, 'Butter', '1 tbsp'),
('e5ff2a13-cc76-5dd2-bda1-53ed68ee5fd7'::uuid, 'Berries', '1 cup'),
('e5ff2a13-cc76-5dd2-bda1-53ed68ee5fd7'::uuid, 'Milk', '1/2 cup');

INSERT INTO ingredients (recipe_id, name, measurement)
SELECT recipe_id, name, measurement
FROM seed_ingredients;

COMMIT;
