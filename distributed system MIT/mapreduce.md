paper for lesson 1:

paper : MapReduce: Simplified Data Processing on Large Clusters
        Jeffrey Dean and Sanjay Ghemawat
        jeff@google.com, sanjay@google.com
        Google, Inc.

mapReduce idea comes from the concept of language lisp's concept: map and reduce, what are them?

map : [Function]

      map result-type function sequence &rest more-sequences

      The function must take as many arguments as there are sequences provided; at least one sequence must be provided.
      The result of map is a sequence such that element j is the result of applying function to element j of each of the
      argument sequences. The result sequence is as long as the shortest of the input sequences.
      -copy from https://www.cs.cmu.edu/Groups/AI/html/cltl/clm/node143.html

      it kind like java lambda expression map.I guess java also learn from this too.
reduce: [Function]

        reduce function sequence &key :from-end :start :end :initial-value

        The reduce function combines all the elements of a sequence using a binary operation; for example, using + one
        can add up all the elements.
        -copy from https://www.cs.cmu.edu/Groups/AI/html/cltl/clm/node143.html

        it just like some assemble operation in java stream.I guess also learn from this.

the process picture:
![pi](mapreduce.png)



