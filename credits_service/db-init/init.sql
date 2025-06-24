--
-- PostgreSQL database dump
--

-- Dumped from database version 14.17 (Ubuntu 14.17-0ubuntu0.22.04.1)
-- Dumped by pg_dump version 14.17 (Ubuntu 14.17-0ubuntu0.22.04.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: credits_inst; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.credits_inst (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    credits integer NOT NULL
);


ALTER TABLE public.credits_inst OWNER TO postgres;

--
-- Name: credits_inst_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.credits_inst_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.credits_inst_id_seq OWNER TO postgres;

--
-- Name: credits_inst_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.credits_inst_id_seq OWNED BY public.credits_inst.id;


--
-- Name: credits_inst id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.credits_inst ALTER COLUMN id SET DEFAULT nextval('public.credits_inst_id_seq'::regclass);


--
-- Data for Name: credits_inst; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.credits_inst (id, name, credits) FROM stdin;
1	NTUA	50
2	EKPA	50
\.


--
-- Name: credits_inst_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.credits_inst_id_seq', 2, true);


--
-- Name: credits_inst credits_inst_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.credits_inst
  ADD CONSTRAINT uq_credits_inst_name UNIQUE (name);



--
-- PostgreSQL database dump complete
--

