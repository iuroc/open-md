package main

import "strings"

// codeSample is one fenced code block. « is replaced with a backtick so a
// sample can contain template literals, struct tags and quasiquotes.
type codeSample struct {
	lang string
	body string
}

func codeDoc(title string, samples ...codeSample) string {
	var b strings.Builder
	b.WriteString("# " + title + "\n\n")
	b.WriteString("每个代码块使用不同语言标记，用来检查语法高亮是否生效。\n\n")
	for _, sample := range samples {
		b.WriteString("```" + sample.lang + "\n")
		b.WriteString(strings.ReplaceAll(sample.body, "«", "`"))
		b.WriteString("\n```\n\n")
	}
	return b.String()
}

func codeFixtures() []Fixture {
	return []Fixture{
		fixture("05a-code-systems.md", codeDoc("代码高亮：系统与编译型语言",
			codeSample{"go", `package main

import (
	"context"
	"fmt"
	"sync"
)

type Number interface{ ~int | ~float64 }

type Queue[T any] struct {
	mu    sync.Mutex
	items []T
	Tags  []string «json:"tags,omitempty"«
}

func Map[T, U any](ctx context.Context, in []T, fn func(context.Context, T) (U, error)) ([]U, error) {
	out := make([]U, 0, len(in))
	for i, v := range in {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("map canceled at %d: %w", i, ctx.Err())
		default:
		}
		u, err := fn(ctx, v)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}
`},
			codeSample{"rust", `use std::collections::HashMap;
use std::fmt::{self, Display};

#[derive(Debug, Clone, PartialEq)]
enum Value<'a> {
    Number(f64),
    Text(&'a str),
}

impl<'a> Display for Value<'a> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Number(n) => write!(f, "{n:.2}"),
            Self::Text(s) => write!(f, "{s:?}"),
        }
    }
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let raw = r#"raw "string" with \n"#;
    let map: HashMap<&str, i64> = [("a", 0x1f), ("b", 0b1010)].into();
    let squares: Vec<_> = (0..10).filter(|n| n % 2 == 0).map(|n| n * n).collect();
    match squares.as_slice() {
        [] => println!("empty"),
        [first, .., last] => println!("{first}..{last} {}", map.len()),
    }
    let _ = raw;
    Ok(())
}
`},
			codeSample{"c", `#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#define SQUARE(x) ((x) * (x))
#define LOG(fmt, ...) fprintf(stderr, "[%s] " fmt "\n", __func__, __VA_ARGS__)

typedef struct node {
    struct node *next;
    uint32_t value;
} node_t;

static int compare(const void *a, const void *b) {
    return (*(const int *)a) - (*(const int *)b);
}

int main(void) {
    node_t *head = NULL;
    for (uint32_t i = 0; i < 0x10; ++i) {
        node_t *next = calloc(1, sizeof *next);
        if (!next) return EXIT_FAILURE;
        next->value = SQUARE(i);
        next->next = head;
        head = next;
    }
    char buf[32];
    snprintf(buf, sizeof buf, "%08x", 0xdeadbeefu);
    LOG("%s", buf);
    return EXIT_SUCCESS;
}
`},
			codeSample{"cpp", `#include <algorithm>
#include <iostream>
#include <memory>
#include <string>
#include <vector>

namespace demo {

template <typename T>
concept Numeric = std::is_arithmetic_v<T>;

template <Numeric T>
class Accumulator {
public:
    explicit Accumulator(T start = T{}) : total_(start) {}
    auto add(T value) -> Accumulator & { total_ += value; return *this; }
    [[nodiscard]] constexpr T total() const noexcept { return total_; }

private:
    T total_;
};

struct Shape {
    virtual ~Shape() = default;
    virtual double area() const = 0;
};

} // namespace demo

int main() {
    auto values = std::vector{1, 2, 3, 4};
    std::ranges::sort(values, std::greater{});
    for (const auto &v : values) std::cout << v << '\n';
    auto ptr = std::make_unique<demo::Accumulator<int>>(42);
    ptr->add(1).add(2);
    return 0;
}
`},
			codeSample{"csharp", `using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;

namespace Demo;

public record Point(double X, double Y)
{
    public double Length => Math.Sqrt(X * X + Y * Y);
}

public interface IRepository<T> where T : class
{
    IEnumerable<T> All();
}

public class Service(IRepository<Point> repo) : IDisposable
{
    public async Task<IReadOnlyList<Point>> LoadAsync(CancellationToken ct = default)
    {
        await Task.Yield();
        var points = repo.All().Where(p => p.Length > 1.0).OrderBy(p => p.X).ToList();
        return points;
    }

    public void Dispose() { }
}

var text = $@"line {42}
raw ""quoted""";
`},
			codeSample{"java", `package demo;

import java.util.List;
import java.util.stream.Collectors;

public sealed interface Shape permits Circle, Square {}
record Circle(double radius) implements Shape {}
record Square(double side) implements Shape {}

public final class Calculator<T extends Number> {
    @SafeVarargs
    public final double sum(T... values) {
        double total = 0;
        for (var value : values) total += value.doubleValue();
        return total;
    }

    public static void main(String[] args) {
        var shapes = List.of(new Circle(1.0), new Square(2.0));
        var areas = shapes.stream()
            .map(s -> switch (s) {
                case Circle c -> Math.PI * c.radius() * c.radius();
                case Square q -> q.side() * q.side();
            })
            .collect(Collectors.toList());
        System.out.printf("%.2f%n", areas.get(0));
        String block = """
                text block
                with "quotes"
                """;
    }
}
`},
			codeSample{"kotlin", `package demo

import kotlinx.coroutines.flow.*

data class User(val id: Long, val name: String, val tags: Set<String> = emptySet())

sealed interface Result<out T> {
    data class Ok<T>(val value: T) : Result<T>
    data class Err(val message: String, val cause: Throwable? = null) : Result<Nothing>
}

suspend fun load(ids: List<Long>): List<User> = flow {
    ids.forEach { id -> emit(User(id, "user-$id")) }
}.filter { it.id % 2L == 0L }.toList()

fun main() {
    val text = """
        raw "string" with ${'$'}money
    """.trimIndent()
    val result = runCatching { error("boom") }
    when (val r = result) {
        is Result.Ok -> println(r.value)
        is Result.Err -> println(r.message)
    }
}
`},
			codeSample{"swift", `import Foundation

protocol Shape {
    var area: Double { get }
}

struct Circle: Shape, Equatable {
    let radius: Double
    var area: Double { .pi * radius * radius }
}

enum NetworkError: Error {
    case timeout(seconds: Int)
    case invalidResponse(status: Int, body: String?)
}

func fetch(_ url: URL) async throws -> Data {
    let (data, response) = try await URLSession.shared.data(from: url)
    guard let http = response as? HTTPURLResponse else {
        throw NetworkError.invalidResponse(status: -1, body: nil)
    }
    return data
}

let raw = #"raw "string" \n"#
let joined = [1, 2, 3].filter { $0.isMultiple(of: 2) }.map { "\($0)" }
`},
			codeSample{"scala", `package demo

import scala.util.{Failure, Success, Try}

sealed trait Tree[+A]
case object Leaf extends Tree[Nothing]
final case class Branch[A](left: Tree[A], value: A, right: Tree[A]) extends Tree[A]

object Demo:
  given Ordering[Branch[Int]] = Ordering.by(_.value)

  extension [A](tree: Tree[A])
    def size: Int = tree match
      case Leaf => 0
      case Branch(l, _, r) => 1 + l.size + r.size

  def parse(input: String): Either[String, Int] =
    input.toIntOption.toRight(s"not a number: $input")

  def main(args: Array[String]): Unit =
    val raw = """raw "string" with ${notInterpolated}"""
    Try(parse("42")) match
      case Success(Right(n)) => println(f"$n%04d")
      case Success(Left(err)) => System.err.println(err)
      case Failure(ex) => ex.printStackTrace()
`},
			codeSample{"nasm", `section .data
    msg db "hello, %s", 10, 0
    value dq 0xDEADBEEF

section .text
    global _start
    extern printf

_start:
    push rbp
    mov rbp, rsp
    xor eax, eax
    mov rdi, msg
    lea rsi, [rel value]
    call printf wrt ..plt
    mov rax, 60
    xor edi, edi
    syscall
`},
		)),

		fixture("05b-code-scripting.md", codeDoc("代码高亮：脚本语言",
			codeSample{"python", `from __future__ import annotations

import asyncio
from dataclasses import dataclass, field
from typing import AsyncIterator, Generic, TypeVar

T = TypeVar("T")

@dataclass(slots=True)
class Item(Generic[T]):
    name: str
    value: T
    tags: list[str] = field(default_factory=list)

    def __post_init__(self) -> None:
        assert self.name, "name required"

async def stream(limit: int) -> AsyncIterator[Item[int]]:
    for index in range(limit):
        await asyncio.sleep(0)
        yield Item(name=f"item-{index}", value=index ** 2,
                   tags=["even"] if index % 2 == 0 else [])

async def main() -> None:
    items = [item async for item in stream(5) if item.value > 1]
    match items:
        case []:
            print("empty")
        case [first, *rest] if (count := len(rest)) > 0:
            print(first.name, count)
    print(f"{items[0].value=:.2f} {items[0].name!r}")
    raw = r"\d+\.\d+"
    doc = """multi-line
    string with 'quotes'
    """
`},
			codeSample{"javascript", `import { EventEmitter } from "node:events";

/** JSDoc block */
export class Store extends EventEmitter {
  #items = new Map();

  constructor({ name = "store", version = "v1" } = {}) {
    super();
    this.name = name;
    this.version = version;
  }

  get size() {
    return this.#items.size;
  }

  async *[Symbol.asyncIterator]() {
    for (const [key, value] of this.#items) {
      await Promise.resolve();
      yield { key, value };
    }
  }

  upsert(key, value) {
    const next = { ...this.#items.get(key), ...value, updatedAt: Date.now() };
    this.#items.set(key, next);
    this.emit("change", { key, next });
    return this;
  }
}

const store = new Store({ version: "v2" });
store.upsert("a", { count: 1 });
const path = «/api/${store.version}/items«;
const pattern = /^\/api\/(?<version>v\d+)\//i;
const label = store.size ?? 0;
`},
			codeSample{"typescript", `type Result<T, E = Error> =
  | { ok: true; value: T }
  | { ok: false; error: E };

interface Identifiable<Id extends string | number = number> {
  readonly id: Id;
  metadata?: Record<string, unknown>;
}

function unwrap<T, E>(result: Result<T, E>): T {
  if (!result.ok) throw result.error;
  return result.value;
}

async function load<T extends Identifiable>(url: URL): Promise<T[]> {
  const response = await fetch(url, { headers: { accept: "application/json" } });
  if (!response.ok) throw new Error(«HTTP ${response.status}«);
  return (await response.json()) as T[];
}

const ids = [1, 2, 3] as const;
type Id = (typeof ids)[number];
const map = new Map<Id, string>(ids.map((id) => [id, «item-${id}«]));
`},
			codeSample{"ruby", `# frozen_string_literal: true

module Demo
  class Store
    include Enumerable

    attr_reader :items

    def initialize(**options)
      @items = options.fetch(:items, {})
      @name = options[:name] || "store"
    end

    def each(&block)
      items.each(&block)
    end

    def upsert(key, **value)
      items[key] = items.fetch(key, {}).merge(value).merge(updated_at: Time.now)
      self
    end

    def summary
      items.map { |key, value| "#{key}=#{value[:count].to_i}" }.join(", ")
    end
  end
end

store = Demo::Store.new(name: "demo", items: { a: { count: 1 } })
store.upsert(:b, count: 2) if store.any? { |_k, v| v[:count].positive? }
puts store.summary
regex = /\A(?<scheme>https?):\/\/(?<host>[^\/]+)/i
`},
			codeSample{"php", `<?php
declare(strict_types=1);

namespace Demo;

use InvalidArgumentException;
use JsonSerializable;

final class Money implements JsonSerializable
{
    public function __construct(
        private readonly int $amount,
        private readonly string $currency = 'CNY',
    ) {
        if ($amount < 0) {
            throw new InvalidArgumentException("negative amount: {$amount}");
        }
    }

    public function add(self $other): self
    {
        if ($other->currency !== $this->currency) {
            throw new InvalidArgumentException('currency mismatch');
        }
        return new self($this->amount + $other->amount, $this->currency);
    }

    public function jsonSerialize(): array
    {
        return ['amount' => $this->amount / 100, 'currency' => $this->currency];
    }
}

$price = new Money(1999);
echo json_encode($price, JSON_THROW_ON_ERROR) . PHP_EOL;
`},
			codeSample{"perl", `#!/usr/bin/env perl
use strict;
use warnings;
use feature 'signatures';
no warnings 'experimental::signatures';

package Demo::Store;

sub new ( $class, %args ) {
    my $self = { items => $args{items} // {}, name => $args{name} // 'store' };
    return bless $self, $class;
}

sub upsert ( $self, $key, %value ) {
    my $entry = { %{ $self->{items}{$key} // {} }, %value, updated_at => time() };
    $self->{items}{$key} = $entry;
    return $self;
}

my $name = 'world';
print "hello, ${name}\n";
my @squares = map { $_ * $_ } grep { $_ % 2 == 0 } 1 .. 10;
my $re = qr/^(?<scheme>https?):\/\/(?<host>.+)$/i;
my $text = <<"END";
heredoc with $name and @squares
END
`},
			codeSample{"lua", `local Store = {}
Store.__index = Store

function Store.new(options)
  local self = setmetatable({}, Store)
  self.name = options.name or "store"
  self.items = options.items or {}
  return self
end

function Store:upsert(key, value)
  local entry = self.items[key] or {}
  for k, v in pairs(value) do
    entry[k] = v
  end
  entry.updated_at = os.time()
  self.items[key] = entry
  return self
end

local function map(tbl, fn)
  local out = {}
  for i, v in ipairs(tbl) do out[i] = fn(v, i) end
  return out
end

local squares = map({ 1, 2, 3, 4 }, function(n) return n ^ 2 end)
local text = [[long bracket string with "quotes" and \n]]
for i, n in ipairs(squares) do
  if n % 2 == 0 then print(("%d -> %d"):format(i, n)) end
end
`},
			codeSample{"bash", `#!/usr/bin/env bash
set -Eeuo pipefail
IFS=$'\n\t'

readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
declare -A counts=([ok]=0 [fail]=0)

log() { printf '[%s] %s\n' "$(date +%H:%M:%S)" "$*" >&2; }

retry() {
  local attempts=$1 delay=$2
  shift 2
  local n=0
  until "$@"; do
    (( ++n >= attempts )) && return 1
    sleep "$delay"
  done
}

for file in "$SCRIPT_DIR"/cases/*.md; do
  [[ -f $file ]] || continue
  if grep -Eq '^#{1,6} ' -- "$file" 2>/dev/null; then
    log "ok: ${file##*/}"
    (( counts[ok]++ ))
  else
    log "fail: $file"
    (( counts[fail]++ )) || true
  fi
done

trap 'log "exit $?"' EXIT
printf 'ok=%d fail=%d\n' "${counts[ok]}" "${counts[fail]}"
`},
			codeSample{"powershell", `#Requires -Version 7.0
[CmdletBinding()]
param(
    [Parameter(Mandatory)]
    [ValidateSet('Light', 'Big')]
    [string] $Mode = 'Light',

    [int] $Limit = 10
)

class CaseFile {
    [string] $Path
    [int] $Size

    CaseFile([string] $path) {
        $this.Path = $path
        $this.Size = (Get-Item -LiteralPath $path).Length
    }

    [string] ToString() { return "$($this.Path) ($($this.Size) bytes)" }
}

$items = Get-ChildItem -Path ./cases -Filter *.md -File |
    Where-Object { $_.Length -gt 0 } |
    ForEach-Object { [CaseFile]::new($_.FullName) } |
    Sort-Object Size -Descending |
    Select-Object -First $Limit

foreach ($item in $items) {
    Write-Host ("{0,-40} {1,10:n0}" -f $item.Path, $item.Size)
}

$here = @"
raw here-string for $Mode
"@
`},
			codeSample{"r", `library(stats)

new_store <- function(name = "store", items = list()) {
  structure(list(name = name, items = items), class = "store")
}

upsert <- function(store, key, ...) {
  stopifnot(is.character(key), length(key) == 1L)
  old <- store$items[[key]] %||% list()
  store$items[[key]] <- modifyList(old, list(...))
  store
}

«%||%« <- function(x, y) if (is.null(x)) y else x

values <- c(1, 2, 3, 4, 5)
squares <- vapply(values[values %% 2 == 0], function(x) x^2, numeric(1))
model <- lm(mpg ~ wt + hp, data = mtcars)
summary(model)$r.squared |> round(3) |> print()
raw <- r"(C:\path\with"quotes")"
`},
			codeSample{"julia", `module Demo

export Store, upsert!

struct Store{T}
    name::String
    items::Dict{String,T}
end

Store(name::String = "store") = Store(name, Dict{String,Any}())

function upsert!(store::Store, key::AbstractString; kwargs...)
    entry = merge(get(store.items, key, Dict()), Dict(kwargs))
    store.items[key] = entry
    return store
end

end # module

using .Demo

store = Store("demo")
upsert!(store, "a"; count = 1, tags = ["x", "y"])
squares = [n^2 for n in 1:10 if iseven(n)]
result = mapreduce(identity, +, squares; init = 0)
println("result = $result")
raw = raw"C:\path\with\slashes"
`},
			codeSample{"dart", `import 'dart:async';
import 'dart:convert';

sealed class Result<T> {
  const Result();
}

final class Ok<T> extends Result<T> {
  const Ok(this.value);
  final T value;
}

final class Err extends Result<Never> {
  const Err(this.message);
  final String message;
}

Future<List<int>> loadSquares(int limit) async* {
  for (var i = 1; i <= limit; i++) {
    await Future<void>.delayed(Duration.zero);
    yield i * i;
  }
}

void main() async {
  final values = await loadSquares(5).where((n) => n.isEven).toList();
  final raw = r'raw \n string';
  final dynamicText = 'value: $values';
  final json = jsonEncode({'values': values, 'ok': true});
  switch (json) {
    case String _ when values.isNotEmpty:
      print(dynamicText);
  }
}
`},
		)),

		fixture("05c-code-functional.md", codeDoc("代码高亮：函数式语言",
			codeSample{"haskell", `{-# LANGUAGE OverloadedStrings #-}
{-# LANGUAGE LambdaCase #-}

module Main (main) where

import Control.Monad (forM_)
import Data.Map.Strict qualified as Map
import Data.Text (Text)

data Shape
  = Circle { radius :: Double }
  | Rectangle { width, height :: Double }
  deriving (Eq, Show)

area :: Shape -> Double
area = \case
  Circle r -> pi * r * r
  Rectangle w h -> w * h

newtype Store = Store (Map.Map Text Int)
  deriving (Semigroup, Monoid) via (Map.Map Text Int)

main :: IO ()
main = do
  let shapes = [Circle 1.0, Rectangle 2.0 3.0]
      total = sum (map area shapes)
      store = Store (Map.fromList [("a", 1), ("b", 2)])
  forM_ shapes $ \shape ->
    putStrLn $ show shape <> " -> " <> show (area shape)
  print total
`},
			codeSample{"elixir", `defmodule Demo.Store do
  @moduledoc """
  A tiny store with GenServer state.
  """

  use GenServer
  require Logger

  @type key :: atom() | String.t()
  @type entry :: %{optional(atom()) => term()}

  @spec start_link(keyword()) :: GenServer.on_start()
  def start_link(opts \\ []), do: GenServer.start_link(__MODULE__, %{}, opts)

  @impl true
  def init(state), do: {:ok, state}

  @impl true
  def handle_cast({:upsert, key, value}, state) do
    updated = Map.update(state, key, value, &Map.merge(&1, value))
    Logger.debug("upsert #{inspect(key)}")
    {:noreply, updated}
  end

  def squares(limit) when is_integer(limit) and limit > 0 do
    for n <- 1..limit, rem(n, 2) == 0, do: n * n
  end
end

store = Demo.Store.start_link!()
GenServer.cast(store, {:upsert, :a, %{count: 1}})
`},
			codeSample{"erlang", `-module(demo_store).
-behaviour(gen_server).

-export([start_link/0, upsert/3, squares/1]).
-export([init/1, handle_call/3, handle_cast/2, handle_info/2]).

-record(state, {items = #{}}).

start_link() ->
    gen_server:start_link({local, ?MODULE}, ?MODULE, [], []).

upsert(Pid, Key, Value) ->
    gen_server:cast(Pid, {upsert, Key, Value}).

squares(Limit) when is_integer(Limit), Limit > 0 ->
    [N * N || N <- lists:seq(1, Limit), N rem 2 =:= 0].

init([]) ->
    {ok, #state{}}.

handle_call(_Request, _From, State) ->
    {reply, ok, State}.

handle_cast({upsert, Key, Value}, #state{items = Items} = State) ->
    Updated = maps:update_with(Key, fun(Old) -> maps:merge(Old, Value) end, Value, Items),
    {noreply, State#state{items = Updated}}.

handle_info(_Info, State) ->
    {noreply, State}.
`},
			codeSample{"clojure", `(ns demo.store
  (:require [clojure.string :as str]))

(defrecord Store [name items])

(defn new-store [name]
  (->Store (or name "store") (atom {})))

(defn upsert! [store key value]
  (swap! (:items store) update key #(merge % value))
  store)

(defn squares [limit]
  (->> (range 1 (inc limit))
       (filter even?)
       (map #(* % %))
       (vec)))

(defn summarize [{:keys [name items]}]
  (let [entries (sort-by key @items)]
    (str/join ", " (map (fn [[k v]] (format "%s=%d" (name k) v)) entries))))

(defn -main [& args]
  (let [store (new-store "demo")
        _ (upsert! store :a {:count 1})
        result (squares 10)]
    (println (summarize store) result)))
`},
			codeSample{"fsharp", `module Demo.Store

open System
open System.Collections.Generic

type Shape =
    | Circle of radius: float
    | Rectangle of width: float * height: float

let area shape =
    match shape with
    | Circle radius -> Math.PI * radius * radius
    | Rectangle (width, height) -> width * height

type Store<'T>() =
    let items = Dictionary<string, 'T>()
    member _.Upsert(key: string, value: 'T) =
        items[key] <- value
        value
    member _.Items = items :> IReadOnlyDictionary<_, _>

let squares limit =
    [ 1..limit ]
    |> List.filter (fun n -> n % 2 = 0)
    |> List.map (fun n -> n * n)

let parsed =
    match Int32.TryParse("42") with
    | true, value -> Ok value
    | _ -> Error "not a number"
`},
			codeSample{"ocaml", `module Store : sig
  type 'a t
  val create : string -> 'a t
  val upsert : 'a t -> string -> 'a -> unit
  val items : 'a t -> (string * 'a) list
end = struct
  type 'a t = { name : string; mutable items : (string * 'a) list }

  let create name = { name; items = [] }

  let upsert store key value =
    store.items <- (key, value) :: List.remove_assoc key store.items

  let items store = store.items
end

type shape = Circle of float | Rectangle of float * float

let area = function
  | Circle radius -> Float.pi *. radius *. radius
  | Rectangle (width, height) -> width *. height

let squares limit =
  List.init limit (fun i -> i + 1)
  |> List.filter (fun n -> n mod 2 = 0)
  |> List.map (fun n -> n * n)

let () =
  let results = squares 10 in
  Printf.printf "%d\n" (List.length results)
`},
			codeSample{"scheme", `#lang racket

(define (make-store name)
  (let ([items (make-hash)])
    (lambda (msg . args)
      (case msg
        [(name) name]
        [(upsert!)
         (match-define (list key value) args)
         (hash-set! items key value)]
        [(items) (hash->list items)]
        [else (error "unknown message" msg)]))))

(define (squares limit)
  (for/list ([n (in-range 1 (add1 limit))]
             #:when (even? n))
    (* n n)))

(define store (make-store "demo"))
(store 'upsert! 'a 42)
(displayln (store 'items))
(displayln (squares 10))
`},
			codeSample{"lisp", `(defpackage :demo.store
  (:use :cl)
  (:export :make-store :upsert! :squares))

(in-package :demo.store)

(defstruct store
  (name "store" :type string)
  (items (make-hash-table :test #'equal) :type hash-table))

(defun upsert! (store key value)
  (setf (gethash key (store-items store)) value)
  store)

(defun squares (limit)
  (loop for n from 1 to limit
        when (evenp n)
          collect (* n n)))

(defmacro with-store ((var &rest options) &body body)
  «(let ((,var (make-store ,@options)))
     ,@body
     ,var))

(with-store (s)
  (upsert! s "a" 1))
(format t "~{~a~^, ~}~%" (squares 10))
`},
		)),

		fixture("05d-code-data-web.md", codeDoc("代码高亮：数据、配置与 Web",
			codeSample{"json", `{
  "name": "openmd-cases",
  "version": 2,
  "enabled": true,
  "limits": { "docs": 500, "bytes": 26214400 },
  "tags": ["markdown", "中文", "emoji 🚀"],
  "nested": { "nullValue": null, "numbers": [0, -1, 3.14, 6.02e23] }
}
`},
			codeSample{"yaml", `%YAML 1.2
---
name: openmd-cases
version: 2
enabled: true
limits: &limits
  docs: 500
  bytes: 26214400
stages:
  - build
  - test
  - deploy
matrix:
  os: [windows, linux]
  include:
    - os: windows
      shell: powershell
defaults:
  <<: *limits
  retries: 3
multiline: |
  first line
  second line
folded: >
  this is
  folded text
anchors:
  - &item {name: alpha, value: 1}
  - *item
`},
			codeSample{"toml", `title = "OpenMD cases"
enabled = true
count = 42
ratio = 3.14
created = 2026-11-21T12:34:56Z
hex = 0xdead_beef

[limits]
docs = 500
bytes = 26_214_400

[owner]
name = "易舒"
tags = ["markdown", "中文"]

[[matrix]]
os = "windows"
shell = "powershell"

[[matrix]]
os = "linux"
shell = "bash"
`},
			codeSample{"sql", `WITH ranked AS (
  SELECT
    d.id,
    d.title,
    COUNT(v.id) AS views,
    ROW_NUMBER() OVER (PARTITION BY d.category ORDER BY COUNT(v.id) DESC) AS rn,
    CASE WHEN COUNT(v.id) > 100 THEN 'hot' ELSE 'normal' END AS state
  FROM documents AS d
  LEFT JOIN views AS v
    ON v.document_id = d.id
   AND v.created_at > NOW() - INTERVAL '7 days'
  WHERE d.deleted_at IS NULL
  GROUP BY d.id, d.title, d.category
)
SELECT id, title, views, state
FROM ranked
WHERE rn <= 3
ORDER BY views DESC
LIMIT 20;
`},
			codeSample{"html", `<!doctype html>
<html lang="zh-CN" data-theme="light">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>OpenMD &amp; fixtures</title>
  <style>
    body { margin: 0; font-family: system-ui, sans-serif; }
  </style>
  <script type="module">
    const state = { count: 0 };
    document.querySelector("#inc")?.addEventListener("click", () => {
      state.count += 1;
      document.querySelector("#out").textContent = «count=${state.count}«;
    });
  </script>
</head>
<body>
  <h1 class="title">标题</h1>
  <button id="inc" aria-label="增加">+</button>
  <output id="out">0</output>
</body>
</html>
`},
			codeSample{"css", `:root {
  --accent: #0f766e;
  --radius: 0.5rem;
}

.card {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
  gap: 1rem;
  padding: 1.5rem;
  border: 1px solid color-mix(in oklab, var(--accent) 25%, transparent);
  border-radius: var(--radius);
}

.card:hover,
.card:focus-visible {
  outline: 2px solid var(--accent);
  transform: translateY(-2px);
}

@media (prefers-color-scheme: dark) {
  :root { --accent: #5eead4; }
}

@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

.card > .title::before {
  content: "→ ";
  animation: fade-in 200ms ease-out both;
}
`},
			codeSample{"scss", `@use "sass:map";
@use "sass:color";

$accent: #0f766e;
$breakpoints: (sm: 640px, md: 768px, lg: 1024px);

@mixin respond($name) {
  $value: map.get($breakpoints, $name);
  @media (min-width: $value) { @content; }
}

%card-base {
  padding: 1rem;
  border-radius: 0.5rem;
}

.card {
  @extend %card-base;
  background: color.adjust($accent, $lightness: 45%);

  &__title {
    font-weight: 700;

    &:hover { color: $accent; }
  }

  @include respond(md) {
    display: grid;
    grid-template-columns: 1fr 2fr;
  }
}
`},
			codeSample{"xml", `<?xml version="1.0" encoding="UTF-8"?>
<catalog xmlns="urn:openmd:cases"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <item id="a" enabled="true">
    <name>Alpha</name>
    <tags>
      <tag>markdown</tag>
      <tag>中文</tag>
    </tags>
    <description><![CDATA[Raw <b>text</b> & entities]]></description>
  </item>
  <!-- comment -->
  <empty />
</catalog>
`},
			codeSample{"graphql", `"""文档节点"""
type Document {
  id: ID!
  title: String!
  tags: [String!]!
  views: Int @deprecated(reason: "use stats")
}

interface Node {
  id: ID!
}

input DocumentFilter {
  q: String
  tags: [String!]
  first: Int = 20
}

enum Order {
  ASC
  DESC
}

type Query {
  document(id: ID!): Document
  search(filter: DocumentFilter): [Document!]!
}

query Search($filter: DocumentFilter) {
  search(filter: $filter) {
    ...DocFields
  }
}

fragment DocFields on Document {
  id
  title
  tags
}
`},
			codeSample{"protobuf", `syntax = "proto3";

package openmd.cases.v1;

option go_package = "openmd/cases/v1;casesv1";

message Document {
  string id = 1;
  string title = 2;
  repeated string tags = 3;
  map<string, int64> counters = 4;
  google.protobuf.Timestamp created_at = 5;

  oneof source {
    string path = 6;
    bytes inline = 7;
  }
}

enum State {
  STATE_UNSPECIFIED = 0;
  STATE_READY = 1;
  STATE_FAILED = 2;
}

service DocumentService {
  rpc Get(GetRequest) returns (Document);
  rpc Watch(GetRequest) returns (stream Document);
}
`},
			codeSample{"ini", `; comment
[server]
host = 127.0.0.1
port = 8080
enabled = true

[paths]
root = C:\openmd\cases
pattern = *.md

[limits]
docs = 500
bytes = 26214400

[owner]
name = 易舒
`},
			codeSample{"diff", `diff --git a/README.md b/README.md
index 1111111..2222222 100644
--- a/README.md
+++ b/README.md
@@ -1,5 +1,6 @@
 # OpenMD

+新增一行
-旧的一行
 unchanged
+
`},
			codeSample{"dockerfile", `# syntax=docker/dockerfile:1.7
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/openmd .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/openmd /usr/local/bin/openmd
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/openmd"]
EXPOSE 8080
`},
			codeSample{"makefile", `GO ?= go
BIN := openmd.exe
CASES := cases

.PHONY: all test cases big clean

all: test $(BIN)

test:
	$(GO) test ./...

cases:
	$(GO) run ./tools/cases

big:
	$(GO) run ./tools/cases -big -force

$(BIN):
	$(GO) build -trimpath -ldflags "-s -w" -o $@ .

clean:
	rm -rf $(CASES) $(BIN)
`},
			codeSample{"nginx", `user nginx;
worker_processes auto;

events { worker_connections 1024; }

http {
  include mime.types;

  server {
    listen 443 ssl http2;
    server_name example.com;
    root /srv/openmd;

    location ~* \.(md|markdown)$ {
      add_header Content-Type "text/plain; charset=utf-8" always;
      try_files $uri =404;
    }

    location /api/ {
      proxy_pass http://127.0.0.1:8080;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
    }
  }
}
`},
			codeSample{"terraform", `terraform {
  required_version = ">= 1.8"
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5"
    }
  }
}

variable "cases" {
  type = map(object({ docs = number, bytes = number }))
  default = {
    light = { docs = 500, bytes = 26214400 }
  }
}

resource "local_file" "checklist" {
  filename = "${path.module}/cases/CHECKLIST.md"
  content  = <<-EOT
    # OpenMD
    mode = light
  EOT
}

output "path" {
  value = local_file.checklist.filename
}
`},
			codeSample{"latex", `\documentclass[11pt,a4paper]{article}
\usepackage{amsmath}
\usepackage{hyperref}

\newcommand{\norm}[1]{\left\lVert #1 \right\rVert}

\begin{document}
\section{Introduction}
Let $x \in \mathbb{R}^n$ and
\begin{equation}
  \norm{x}_2 = \sqrt{\sum_{i=1}^{n} x_i^2}.
\end{equation}
\href{https://example.com}{OpenMD} uses \textbf{Markdown}.
\end{document}
`},
		)),
	}
}
