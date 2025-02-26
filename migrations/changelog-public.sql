
--changeset solomon.public:1 labels:public context:public
--comment: tier_time definition
CREATE TYPE "public"."tier_times" AS ENUM ('low', 'medium', 'high', 'batch');
--rollback DROP TYPE "public"."tier_times";

--changeset solomon.public:2 labels:public context:public
--comment: tier_usage definition
CREATE TYPE "public"."tier_usages" AS ENUM ('low', 'medium', 'high');
--rollback DROP TYPE "public"."tier_usages";

--changeset solomon.public:3 labels:public context:public
--comment: tier_models definition
CREATE TYPE "public"."tier_models" AS ENUM ('low', 'medium', 'high');
--rollback DROP TYPE "public"."tier_models";
