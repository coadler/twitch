--
-- PostgreSQL database dump
--

-- Dumped from database version 10.5
-- Dumped by pg_dump version 10.5

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: games; Type: TABLE; Schema: public; Owner: colin
--

CREATE TABLE public.games (
    id integer NOT NULL,
    name text NOT NULL,
    box_art_url text NOT NULL
);


ALTER TABLE public.games OWNER TO colin;

--
-- Name: games_id_seq; Type: SEQUENCE; Schema: public; Owner: colin
--

CREATE SEQUENCE public.games_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.games_id_seq OWNER TO colin;

--
-- Name: games_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: colin
--

ALTER SEQUENCE public.games_id_seq OWNED BY public.games.id;


--
-- Name: twitch_user; Type: TABLE; Schema: public; Owner: colin
--

CREATE TABLE public.twitch_user (
    id text NOT NULL,
    login text NOT NULL,
    display_name text NOT NULL,
    type text NOT NULL,
    broadcaster_type text NOT NULL,
    description text NOT NULL,
    profile_image_url text NOT NULL,
    offline_image_url text NOT NULL,
    view_count integer NOT NULL
);


ALTER TABLE public.twitch_user OWNER TO colin;

--
-- Name: twitch_user_id_seq; Type: SEQUENCE; Schema: public; Owner: colin
--

CREATE SEQUENCE public.twitch_user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.twitch_user_id_seq OWNER TO colin;

--
-- Name: twitch_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: colin
--

ALTER SEQUENCE public.twitch_user_id_seq OWNED BY public.twitch_user.id;


--
-- Name: webhooks; Type: TABLE; Schema: public; Owner: colin
--

CREATE TABLE public.webhooks (
    id text NOT NULL,
    token text NOT NULL
);


ALTER TABLE public.webhooks OWNER TO colin;

--
-- Name: webhooks_id_seq; Type: SEQUENCE; Schema: public; Owner: colin
--

CREATE SEQUENCE public.webhooks_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.webhooks_id_seq OWNER TO colin;

--
-- Name: webhooks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: colin
--

ALTER SEQUENCE public.webhooks_id_seq OWNED BY public.webhooks.id;


--
-- Data for Name: games; Type: TABLE DATA; Schema: public; Owner: colin
--

COPY public.games (id, name, box_art_url) FROM stdin;
\.


--
-- Data for Name: twitch_user; Type: TABLE DATA; Schema: public; Owner: colin
--

COPY public.twitch_user (id, login, display_name, type, broadcaster_type, description, profile_image_url, offline_image_url, view_count) FROM stdin;
\.


--
-- Data for Name: webhooks; Type: TABLE DATA; Schema: public; Owner: colin
--

COPY public.webhooks (id, token) FROM stdin;
\.


--
-- Name: games_id_seq; Type: SEQUENCE SET; Schema: public; Owner: colin
--

SELECT pg_catalog.setval('public.games_id_seq', 1, false);


--
-- Name: twitch_user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: colin
--

SELECT pg_catalog.setval('public.twitch_user_id_seq', 1, false);


--
-- Name: webhooks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: colin
--

SELECT pg_catalog.setval('public.webhooks_id_seq', 1, false);


--
-- Name: games games_pkey; Type: CONSTRAINT; Schema: public; Owner: colin
--

ALTER TABLE ONLY public.games
    ADD CONSTRAINT games_pkey PRIMARY KEY (id);


--
-- Name: twitch_user twitch_user_pkey; Type: CONSTRAINT; Schema: public; Owner: colin
--

ALTER TABLE ONLY public.twitch_user
    ADD CONSTRAINT twitch_user_pkey PRIMARY KEY (id);


--
-- Name: webhooks webhooks_pkey; Type: CONSTRAINT; Schema: public; Owner: colin
--

ALTER TABLE ONLY public.webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--

